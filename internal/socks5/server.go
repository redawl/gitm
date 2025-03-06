package socks5

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/redawl/gitm/internal/config"
)

func StartTransparentSocksProxy(conf config.Config) (error) {
    listener, err := net.Listen("tcp", conf.SocksListenUri)

    if err != nil {
        return err
    }

    go func () {
        for {
            client, err := listener.Accept()
            if err != nil {
                slog.Error("Error accepting connection", "error", err)
            }
            
            if err := handleConnection(client, conf); err != nil {
                slog.Error("Error handling connection", "error", err)
            }
        }
    }()

    return nil
}

func handleConnection(client net.Conn, conf config.Config) (error) {
    slog.Debug("Received connection", "address", client.RemoteAddr())

    greeting, err := ParseClientGreeting(client)

    if err != nil {
        return err
    }

    slog.Debug("Parsed client greeting", "greeting", greeting)
    
    if greeting.CanHandle() {
        slog.Debug("Handling Request")
        client.Write(
            FormatServerChoice(SOCKS_VER_5, METHOD_NO_AUTH_REQUIRED),
        )

        request, status := ParseClientConnRequest(client)

        if status != STATUS_SUCCEEDED {
            client.Write(FormatConnResponse(
                SOCKS_VER_5,
                status,
                client.LocalAddr(),
            ))
            return fmt.Errorf("Error parsing client connection request: %d", status)
        }
        
        slog.Debug("Parsed conn request", "request", request)

        if request.DstPort == 80 {
            server, err := net.Dial("tcp", conf.HttpListenUri)
            if err != nil {
                slog.Error("Error contacting http proxy server", "error", err)
                client.Write(FormatConnResponse(
                    SOCKS_VER_5,
                    STATUS_HOST_UNREACHABLE,
                    server.LocalAddr(),
                ))
                return err
            }

            slog.Debug("Proxy success")

            client.Write(FormatConnResponse(
                SOCKS_VER_5,
                STATUS_SUCCEEDED,
                server.LocalAddr(),
            ))

            transparentProxy(client, server)

        } else if request.DstPort == 443 {
            server, err := net.Dial("tcp", conf.TlsListenUri)
            if err != nil {
                slog.Error("Error contacting https proxy server", "error", err)
                client.Write(FormatConnResponse(
                    SOCKS_VER_5,
                    STATUS_HOST_UNREACHABLE,
                    server.LocalAddr(),
                ))
                return err
            }

            slog.Debug("Proxy success")
            client.Write(FormatConnResponse(
                SOCKS_VER_5,
                STATUS_SUCCEEDED,
                server.LocalAddr(),
            ))

            transparentProxy(client, server)
        } else {
            slog.Error("Unrecognized port, forwarding without logging", "request", request)
            server, err := net.Dial("tcp", fmt.Sprintf("%s:%d", request.DstIp, request.DstPort))
            if err != nil {
                client.Write(FormatConnResponse(
                    SOCKS_VER_5,
                    STATUS_HOST_UNREACHABLE,
                    client.RemoteAddr(),
                ))
                return err
            }

            slog.Debug("Proxy success")
            client.Write(FormatConnResponse(
                SOCKS_VER_5,
                STATUS_SUCCEEDED,
                server.LocalAddr(),
            ))

            transparentProxy(client, server)
        }
        
        slog.Debug("Finished proxying request")
    } else {
        slog.Debug("Cannot handle request")
        client.Write(FormatServerChoice(SOCKS_VER_5, METHOD_NO_ACCEPTABLE_METHODS))
    }

    return nil
}

func transparentProxy (client net.Conn, server net.Conn) {
    go io.Copy(client, server)
    go io.Copy(server, client)
}

