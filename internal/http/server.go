package http

import (
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/redawl/gitm/internal/cacert"
	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/db"
	"github.com/redawl/gitm/internal/packet"
)

func ListenAndServe(conf config.Config, httpPacketHandler func(packet.HttpPacket)) (*http.Server, error) {
	server := &http.Server{Addr: conf.HttpListenUri, Handler: Handler(httpPacketHandler, &conf)}

	if ln, err := net.Listen("tcp", conf.HttpListenUri); err != nil {
		return nil, err
	} else {
		go func() {
			if err := server.Serve(ln); errors.Is(err, http.ErrServerClosed) {
				slog.Error("Error starting server", "error", err)
			}
		}()
	}

	return server, nil
}

func ListenAndServeTls(conf config.Config, httpPacketHandler func(packet.HttpPacket)) (*http.Server, error) {
	// Disables certificate checking globally
	// TODO: Should we be sharing this config below?
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	cfg := &tls.Config{
		// Make sure we can forward ALL tls traffic
		// (or as much as possible with go)
		MinVersion: tls.VersionTLS10,
		// If client doesn't care about verifying, neither do we
		InsecureSkipVerify: true,
	}

	cfg.GetCertificate = func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
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
	}

	server := &http.Server{
		Addr:      conf.TlsListenUri,
		Handler:   Handler(httpPacketHandler, &conf),
		TLSConfig: cfg,
	}

	if ln, err := net.Listen("tcp", conf.TlsListenUri); err != nil {
		return nil, err
	} else {
		go func() {
			if err := server.ServeTLS(ln, "", ""); !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Error starting server", "error", err)
			}
		}()
	}

	return server, nil
}
