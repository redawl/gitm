package socks5

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/textproto"
	"os"

	"github.com/redawl/gitm/internal/cacert"
	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/db"
	"github.com/redawl/gitm/internal/httputils"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/util"
)

var (
	CLIENT_CONFIG = &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS10}
	SERVER_CONFIG = &tls.Config{
		// Make sure we can forward ALL tls traffic
		// (or as much as possible with go)
		MinVersion: tls.VersionTLS10,
		// If client doesn't care about verifying, neither do we
		InsecureSkipVerify: true,
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			domainInfo, err := db.GetDomain(chi.ServerName)
			if err != nil {
				return nil, err
			}

			if domainInfo == nil {
				err = cacert.AddHostname(chi.ServerName)
				if err != nil {
					return nil, err
				}

				domainInfo, err = db.GetDomain(chi.ServerName)
				if err != nil {
					return nil, err
				}
			}

			certificate, err := tls.X509KeyPair(domainInfo.Cert, domainInfo.PrivKey)
			if err != nil {
				return nil, err
			}

			return &certificate, nil
		},
	}
)

// ListenAndServeSocks5 starts a socks5 proxy which will pass any intercepted packets to packetHandler.
// If net.Listen fails for the server, an error is returned.
func ListenAndServeSocks5(conf config.Config, packetHandler func(packet.Packet)) (net.Listener, error) {
	if listener, err := net.Listen("tcp", conf.SocksListenUri); err != nil {
		return nil, err
	} else {
		go func() {
			for {
				if client, err := listener.Accept(); err != nil {
					if !errors.Is(err, net.ErrClosed) {
						slog.Error("Error accepting connection", "error", err)
					}
				} else {
					logger := slog.With("RemoteAddr", client.RemoteAddr(), "LocalAddr", client.LocalAddr())
					go func() {
						if err := handleConnection(client, packetHandler); err != nil {
							logger.Error("Error handling connection", "error", err)
						}
					}()
				}
			}
		}()

		return listener, err
	}
}

func handleConnection(client net.Conn, packetHandler func(packet.Packet)) error {
	logger := slog.With("RemoteAddr", client.RemoteAddr(), "LocalAddr", client.LocalAddr())
	logger.Debug("Handling socks5 connection")

	greeting, err := ParseClientGreeting(client)
	if err != nil {
		return fmt.Errorf("parsing client greeting: %w", err)
	}

	logger.Debug("Parsed client greeting", "greeting", greeting)

	if greeting.CanHandle() {
		logger.Debug("Handling Request")
		if _, err := client.Write(
			FormatServerChoice(SOCKS_VER_5, METHOD_NO_AUTH_REQUIRED),
		); err != nil {
			return fmt.Errorf("formatting server choice: %w", err)
		}

		request, status, err := ParseClientConnRequest(client)

		if status != STATUS_SUCCEEDED {
			if _, err := client.Write(FormatConnResponse(
				SOCKS_VER_5,
				status,
				client.LocalAddr(),
			)); err != nil {
				return fmt.Errorf("formatting conn response: %w", err)
			}
			return fmt.Errorf("parsing client connection request: %w", err)
		}

		if request.Cmd == CMD_UDP_ASSOCIATE {
			return handleUdp(client, request)
		}

		logger = logger.With("DstIp", request.DstIp, "DstPort", request.DstPort)

		logger.Debug("Parsed conn request", "request", request)

		if request.DstIp == "gitm" {
			logger.Debug("Handling with gitm webserver")
			if _, err := client.Write(FormatConnResponse(
				SOCKS_VER_5,
				STATUS_SUCCEEDED,
				client.RemoteAddr(),
			)); err != nil {
				return fmt.Errorf("formatting conn response: %w", err)
			}
			return handleGitm(client)
		}

		switch request.DstPort {
		case 80:
			server, err := net.Dial("tcp", net.JoinHostPort(request.DstIp, fmt.Sprintf("%d", request.DstPort)))
			if err != nil {
				logger.Error("Error contacting proxied ip", "error", err)
				if _, err := client.Write(FormatConnResponse(
					SOCKS_VER_5,
					STATUS_HOST_UNREACHABLE,
					client.RemoteAddr(),
				)); err != nil {
					return fmt.Errorf("formatting conn response: %w", err)
				}
				return err
			}
			defer func() {
				if err := server.Close(); err != nil {
					slog.Error("Error closing server", "error", err)
				}
			}()

			logger.Debug("Proxy success")

			if _, err := client.Write(FormatConnResponse(
				SOCKS_VER_5,
				STATUS_SUCCEEDED,
				server.LocalAddr(),
			)); err != nil {
				return fmt.Errorf("formatting conn response: %w", err)
			}

			return httputils.HandleHttpRequest(client, server, packetHandler)
		case 443:
			outboundConn, err := net.Dial("tcp", net.JoinHostPort(request.DstIp, fmt.Sprintf("%d", request.DstPort)))
			if err != nil {
				logger.Error("Error contacting proxied ip", "error", err)
				if _, err := client.Write(FormatConnResponse(
					SOCKS_VER_5,
					STATUS_HOST_UNREACHABLE,
					client.LocalAddr(),
				)); err != nil {
					return fmt.Errorf("sending host unreachable: %w", err)
				}
				return fmt.Errorf("contacting proxied ip: %w", err)
			}
			defer func() {
				if err := outboundConn.Close(); err != nil {
					logger.Error("Error closing outboundConn", "error", err)
				}
			}()

			logger.Debug("Proxy success")
			if _, err := client.Write(FormatConnResponse(
				SOCKS_VER_5,
				STATUS_SUCCEEDED,
				client.RemoteAddr(),
			)); err != nil {
				return err
			}
			inboundConn := tls.Server(client, SERVER_CONFIG)
			defer func() {
				if err := inboundConn.Close(); err != nil {
					logger.Error("Error closing inboundConn", "error", err)
				}
			}()
			if err := inboundConn.Handshake(); err != nil {
				if errors.Is(err, io.EOF) || err.Error() == "tls: client using inappropriate protocol fallback" {
					return nil
				}
				return fmt.Errorf("tls client handshake: %w", err)
			}
			config := CLIENT_CONFIG.Clone()
			config.InsecureSkipVerify = true
			config.ServerName = inboundConn.ConnectionState().ServerName
			return httputils.HandleHttpRequest(inboundConn, tls.Client(outboundConn, config), packetHandler)
		default:
			logger.Info("Unrecognized port, forwarding without logging", "request", request)
			server, err := net.Dial("tcp", net.JoinHostPort(request.DstIp, fmt.Sprintf("%d", request.DstPort)))
			if err != nil {
				_, _ = client.Write(FormatConnResponse(
					SOCKS_VER_5,
					STATUS_HOST_UNREACHABLE,
					client.RemoteAddr(),
				))
				return err
			}

			logger.Debug("Proxy success")
			if _, err := client.Write(FormatConnResponse(
				SOCKS_VER_5,
				STATUS_SUCCEEDED,
				server.LocalAddr(),
			)); err != nil {
				return err
			}
			transparentProxy(client, server)
		}

		logger.Debug("Finished proxying request")
	} else {
		logger.Debug("Cannot handle request")
		if _, err := client.Write(FormatServerChoice(SOCKS_VER_5, METHOD_NO_ACCEPTABLE_METHODS)); err != nil {
			return fmt.Errorf("sending no acceptable methods: %w", err)
		}
	}

	return nil
}

// transparentProxy simply forwards all traffic from client -> server, and vice versa.
// Use transparentProxy when you don't know how to tell when a network packet ends or begins,
// and you don't care about logging the traffic
func transparentProxy(client net.Conn, server net.Conn) {
	logger := slog.With("RemoteAddr", client.RemoteAddr(), "LocalAddr", client.LocalAddr())
	go func() {
		if _, err := io.Copy(client, server); err != nil {
			logger.Error("Error proxying server to client", "error", err)
		}
	}()
	go func() {
		if _, err := io.Copy(server, client); err != nil {
			logger.Error("Error proxying client to server", "error", err)
		}
	}()
}

func handleGitm(client net.Conn) error {
	reader := textproto.NewReader(bufio.NewReader(client))
	_, uri, _, err := httputils.ReadLine1(reader)
	if err != nil {
		return err
	}

	if _, err := reader.ReadMIMEHeader(); err != nil {
		return err
	}

	if uri == "/ca.crt" {
		configDir, err := util.GetConfigDir()
		if err != nil {
			return fmt.Errorf("getting configdir: %w", err)
		}

		certLocation := configDir + "/ca.crt"
		contents, err := os.ReadFile(certLocation)
		if err != nil {
			slog.Error("Error getting ca cert", "error", err)
			if _, err := client.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")); err != nil {
				return fmt.Errorf("sending internal server error: %w", err)
			}
			return nil
		}

		if _, err := fmt.Fprintf(client, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\n", len(contents)); err != nil {
			return fmt.Errorf("sending 200 OK: %w", err)
		}
		if _, err := client.Write(contents); err != nil {
			return fmt.Errorf("sending body: %w", err)
		}
	}

	return nil
}

func ListenAndServePac(conf *config.Config) error {
	return http.ListenAndServe(conf.PacListenUri, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("Handling pac file request")
		if r.URL.Path == "/proxy.pac" {
			_, _ = fmt.Fprintf(w, "function FindProxyForURL(url, host){return \"SOCKS %s\";}", conf.SocksListenUri)
		}
	}))
}
