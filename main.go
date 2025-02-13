package main

import (
	"log/slog"
	"os"

	"com.github.redawl.mitmproxy/cacert"
	"com.github.redawl.mitmproxy/config"
	"com.github.redawl.mitmproxy/http"
	"com.github.redawl.mitmproxy/packet"
	"com.github.redawl.mitmproxy/socks5"
	"com.github.redawl.mitmproxy/ui"
)

func setupbackend (httpHandler func(packet.HttpPacket), httpsHandler func(packet.HttpPacket)) {
    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))

    slog.SetDefault(logger)

    conf := config.Config{
        HttpListenUri: "127.0.0.1:8080",
        TlsListenUri: "127.0.0.1:8443",
    }

    slog.Info("Starting Cacert server")
    go cacert.ListenAndServe("0.0.0.0:9090")

    slog.Info("Starting http server")
    go http.ListenAndServe(conf, httpHandler)
    
    slog.Info("Starting https server")
    go http.ListenAndServeTls(conf, httpsHandler)
    
    slog.Info("Starting socks5 proxy")
    go socks5.StartTransparentSocksProxy("0.0.0.0:1080", conf)
}

func main() {
    packetChan := make(chan packet.HttpPacket)
    setupbackend(func(p packet.HttpPacket){
        packetChan <- p
    }, 
    func(p packet.HttpPacket){
        packetChan <- p
    })
    
    ui.ShowAndRun(packetChan)
}

