package main

import (
	"log/slog"
	"os"

	"github.com/redawl/gitm/internal/cacert"
	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/http"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/socks5"
	"github.com/redawl/gitm/internal/ui"
)

func setupbackend (httpHandler func(packet.HttpPacket)) {
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
    go cacert.ListenAndServe(conf.CacertListenUri, conf.SocksListenUri)

    slog.Info("Starting http server")
    go http.ListenAndServe(conf, httpHandler)
    
    slog.Info("Starting https server")
    go http.ListenAndServeTls(conf, httpHandler)
    
    if err := socks5.StartTransparentSocksProxy(conf); err != nil {
        slog.Info("Started socks5 proxy")
    }
}

func main() {
    packetChan := make(chan packet.HttpPacket)
    setupbackend(func(p packet.HttpPacket){
        packetChan <- p
    })
    
    ui.ShowAndRun(packetChan)
}

