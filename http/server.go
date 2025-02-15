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

func ListenAndServe(conf config.Config, httpPacketHandler func(packet.HttpPacket)) error {
    err := http.ListenAndServe(conf.HttpListenUri, Handler(conf, httpPacketHandler))

    slog.Error("Error serving http proxy server", "error", err)

    return err
}

func ListenAndServeTls(conf config.Config, httpPacketHandler func(packet.HttpPacket)) {
    hostnameInfos, err := db.GetDomains()

    if err != nil {
        slog.Error("Error getting domains", "error", err)
        return
    }

    cfg := &tls.Config{
        MinVersion: 1.0,
    }

    certMap := make(map[string]*tls.Certificate, len(hostnameInfos))

    for _, hostnameInfo := range hostnameInfos {
        cert, err := tls.X509KeyPair(hostnameInfo.Cert, hostnameInfo.PrivKey)

        if err != nil {
            slog.Error("Error loading x509 keypair", "error", err)
            return
        }

        certMap[hostnameInfo.Domain] = &cert
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
        Handler: Handler(conf, httpPacketHandler),
        TLSConfig: cfg,
    }

    err = server.ListenAndServeTLS("", "")

    slog.Error("Error serving https proxy server", "error", err)
}

