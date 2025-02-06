package main

import (
	"log/slog"
	"os"

	"com.github.redawl.mitmproxy/cacert"
	"com.github.redawl.mitmproxy/config"
	"com.github.redawl.mitmproxy/http"
	"com.github.redawl.mitmproxy/packet"
	"com.github.redawl.mitmproxy/socks5"
)

func main () {
    conf := config.Config{
        HttpListenUri: "127.0.0.1:8080",
        TlsListenUri: "127.0.0.1:8443",
    }
    
    // Start ca.crt server
    userCfgDir, err := os.UserConfigDir()
    if err != nil {
        slog.Error("Error getting config dir", "error", err)
        return
    }
    configDir := userCfgDir + "/mitmproxy"
    err = cacert.SetupCertificateAuthority("example.com", configDir + "/server.pem", configDir + "/privkey.pem")

    go cacert.ListenAndServe("0.0.0.0:9090")

    go http.ListenAndServe(conf, func(p packet.Packet) {
        slog.Info("Received packet", "packet", p)
    })

    go http.ListenAndServeTls(conf, func(p packet.Packet) {
        slog.Info("Received tls packet", "packet", p)
    })

    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))

    slog.SetDefault(logger)

    socks5.StartTransparentSocksProxy("127.0.0.1:1080", conf)
}
