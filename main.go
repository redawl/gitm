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
    conf := config.ParseFlags()

    logLevel := slog.LevelInfo

    if conf.Debug {
        logLevel = slog.LevelDebug
    }

    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: logLevel,
    }))

    slog.SetDefault(logger)

    slog.Info("Starting Cacert server")
    go cacert.ListenAndServe(conf.CacertListenUri)

    slog.Info("Starting http server")
    go http.ListenAndServe(conf, httpHandler)
    
    slog.Info("Starting https server")
    go http.ListenAndServeTls(conf, httpsHandler)
    
    slog.Info("Starting socks5 proxy")
    go socks5.StartTransparentSocksProxy(conf)
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

