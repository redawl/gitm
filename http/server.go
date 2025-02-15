package http

import (
	"crypto/tls"
	"log/slog"
	"net/http"

	"com.github.redawl.mitmproxy/cacert"
	"com.github.redawl.mitmproxy/config"
	"com.github.redawl.mitmproxy/db"
	"com.github.redawl.mitmproxy/packet"
)

func ListenAndServe(conf config.Config, httpPacketHandler func(packet.HttpPacket)) {
    slog.Error("Error serving http proxy server", "error", http.ListenAndServe(conf.HttpListenUri, Handler(httpPacketHandler)))
}

func ListenAndServeTls(conf config.Config, httpPacketHandler func(packet.HttpPacket)) {
    cfg := &tls.Config{
        MinVersion: 1.0,
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
        Handler: Handler(httpPacketHandler),
        TLSConfig: cfg,
    }

    slog.Error("Error serving https proxy server", "error", server.ListenAndServeTLS("", ""))
}

