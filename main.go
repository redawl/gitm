package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"com.github.redawl.mitmproxy/packet"
	"com.github.redawl.mitmproxy/socks5"
	"com.github.redawl.mitmproxy/tls"
)

type Args struct {
    ListenUri string
}

func main () {
    args := &Args{}

    flag.StringVar(&args.ListenUri, "l", "0.0.0.0:8080", "ip:port to bind to")

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
        if len(p.Data) == 0 {
            return 
        }

        if p.Data[0] >= 0x14 && p.Data[0] <= 0x18 {
            records := tls.ParseTLSRecords(p.Data)
            slog.Info("Found tls records", "count", len(records))
            for _, record := range(records) {
                attrs := []slog.Attr{
                    slog.String("src", p.SrcIp),
                    slog.String("dst", p.DstIp),
                }
                attrs = append(attrs, record.LogAttrs()...)
                logger := slog.New(slog.Default().Handler().WithAttrs(attrs)) 
                logger.Info("Packet")
            }
        } else {
            slog.Info("Packet", "src", p.SrcIp, "dst", p.DstIp, "data", string(p.Data))
        }

        if err := p.WritePacket(*outFile); err != nil {
            slog.Error("Couldn't write packet to outfile", "error", err)
        }
    })

}
