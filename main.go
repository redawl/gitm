package main

import (
	"fmt"
	"log/slog"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/redawl/gitm/internal/cacert"
	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/http"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/socks5"
	"github.com/redawl/gitm/internal/ui"
)

// setupbackend sets up the socks5 proxy, as well as the http and https proxy listeners.
// Returns a cleanup function for gracefully shutting down the backend
func setupbackend(conf config.Config, httpHandler func(packet.HttpPacket)) (func(), error) {
	logLevel := slog.LevelInfo

	if conf.Debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)
	// Init cacert if it hasn't been already
	err := cacert.InitCaCert()

	if err != nil {
		return nil, err
	}

	httpListener, err := http.ListenAndServe(conf, httpHandler)

	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	httpsListener, err := http.ListenAndServeTls(conf, httpHandler)

	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	socksListener, err := socks5.StartTransparentSocksProxy(conf)

	if err != nil {
		return nil, fmt.Errorf("socks5 proxy: %w", err)
	}

	return func() {
		_ = httpListener.Close()
		_ = httpsListener.Close()
		_ = socksListener.Close()
	}, nil
}

func main() {
	app := app.NewWithID("com.github.redawl.gitm")

	conf := config.ParseFlags(app.Preferences())

	packetChan := make(chan packet.HttpPacket)
	slog.Info("Setting up backend...")
	restart, err := setupbackend(conf, func(p packet.HttpPacket) {
		packetChan <- p
	})

	if err != nil {
		slog.Error("Error setting up backend", "error", err)
		return
	}

	ui.ShowAndRun(app, packetChan, func() {
		restart()
		restart, err = setupbackend(config.ParseFlags(app.Preferences()), func(p packet.HttpPacket) {
			packetChan <- p
		})
	})
}
