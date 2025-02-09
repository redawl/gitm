package main

import (
	"log/slog"
	"os"
	"slices"

	"com.github.redawl.mitmproxy/cacert"
	"com.github.redawl.mitmproxy/config"
	"com.github.redawl.mitmproxy/db"
	"com.github.redawl.mitmproxy/http"
	"com.github.redawl.mitmproxy/packet"
	"com.github.redawl.mitmproxy/socks5"
)

func main () {
    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))

    slog.SetDefault(logger)

    conf := config.Config{
        HttpListenUri: "127.0.0.1:8080",
        TlsListenUri: "127.0.0.1:8443",
    }

    domains, err := db.GetDomains()

    if err != nil {
        slog.Error("Error getting domains", "error", err)
        return
    }

    if !slices.Contains(domains, "example.com") {
        if err := cacert.AddHostname("example.com"); err != nil {
            slog.Error("Error adding domain", "error", err)
            return
        }
    }
    slog.Info("Starting Cacert server")
    go cacert.ListenAndServe("0.0.0.0:9090")

    slog.Info("Starting http server")
    go http.ListenAndServe(conf, func(p packet.Packet) {
        slog.Info("Received packet", "packet", p)
    })
    
    slog.Info("Starting https server")
    go http.ListenAndServeTls(conf, func(p packet.Packet) {
        slog.Info("Received tls packet", "packet", p)
    })

    
    slog.Info("Starting socks5 proxy")
    err = socks5.StartTransparentSocksProxy("127.0.0.1:1080", conf)

    if err != nil {
        slog.Error("Error serving socks proxy", "error", err)
    }
}
