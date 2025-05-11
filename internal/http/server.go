package http

import (
	"crypto/tls"
	"log/slog"
	"net/http"

	"github.com/redawl/gitm/internal/cacert"
	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/db"
	"github.com/redawl/gitm/internal/packet"
)

func ListenAndServe(conf config.Config, httpPacketHandler func(packet.HttpPacket)) {
    slog.Error("Error serving http proxy server", "error", http.ListenAndServe(conf.HttpListenUri, Handler(httpPacketHandler, &conf)))
}

func ListenAndServeTls(conf config.Config, httpPacketHandler func(packet.HttpPacket)) {
    // Disables certificate checking globally
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} 

    cfg := &tls.Config{
        // Make sure we can forward ALL tls traffic
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
        Addr: conf.TlsListenUri,
        Handler: Handler(httpPacketHandler, &conf),
        TLSConfig: cfg,
    }

    slog.Error("Error serving https proxy server", "error", server.ListenAndServeTLS("", ""))
}

