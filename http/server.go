package http

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"com.github.redawl.mitmproxy/cacert"
	"com.github.redawl.mitmproxy/config"
	"com.github.redawl.mitmproxy/db"
	"com.github.redawl.mitmproxy/packet"
	"com.github.redawl.mitmproxy/util"
)

func ListenAndServe(conf config.Config, httpPacketHandler func(packet.HttpPacket)) error {
    err := http.ListenAndServe(conf.HttpListenUri, Handler(conf, httpPacketHandler))

    slog.Error("Error serving http proxy server", "error", err)

    return err
}

func ListenAndServeTls(conf config.Config, httpPacketHandler func(packet.HttpPacket)) {
    configDir, err := util.GetConfigDir()
    if err != nil {
        slog.Error("Error getting config dir", "error", err)
        return
    }

    certDir := configDir + "/certs"

    hostnames, err := db.GetDomains()

    if err != nil {
        slog.Error("Error getting domains", "error", err)
        return
    }

    cfg := &tls.Config{
        MinVersion: 1.0,
    }

    certMap := make(map[string]*tls.Certificate, len(hostnames))

    for _, hostname := range hostnames {
        cert, err := tls.LoadX509KeyPair(fmt.Sprintf("%s/%s.pem", certDir, hostname), fmt.Sprintf("%s/%s-priv.pem", certDir, hostname))

        if err != nil {
            slog.Error("Error loading x509 keypair", "error", err)
            return
        }

        certMap[hostname] = &cert
    }

    cfg.GetCertificate = func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
        hostname := chi.ServerName
        cert, found := certMap[chi.ServerName]

        if found {
            return cert, nil
        }

        err := cacert.AddHostname(chi.ServerName)

        if err != nil {
            return nil, err
        }


        certificate, err := tls.LoadX509KeyPair(fmt.Sprintf("%s/%s.pem", certDir, hostname), fmt.Sprintf("%s/%s-priv.pem", certDir, hostname))

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
