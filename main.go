package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"com.github.redawl.mitmproxy/packet"
	"com.github.redawl.mitmproxy/socks5"
)

type Args struct {
    ListenUri string
}

func main () {
    args := &Args{}

    flag.StringVar(&args.ListenUri, "l", "0.0.0.0:1080", "ip:port to bind to")

    flag.Parse()

    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))

    slog.SetDefault(logger)

    slog.Info(fmt.Sprintf("Starting proxy on %s", args.ListenUri))

    outFile, err := os.OpenFile(fmt.Sprintf("%s/mitmproxy.log", os.TempDir()), os.O_RDWR | os.O_CREATE, 0666)

    if err != nil {
        slog.Error("Failed to open logfile", "error", err)
        return
    }

    socks5.StartTransparentSocksProxy(args.ListenUri, func(p packet.Packet) {
        logger := slog.New(slog.Default().Handler().WithAttrs(p.GetLogAttrs())) 
        logger.Info("Packet")

        if err := p.WritePacket(*outFile); err != nil {
            slog.Error("Couldn't write packet to outfile", "error", err)
        }
    })

}
