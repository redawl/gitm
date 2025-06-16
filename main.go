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

// setupBackend sets up the socks5 proxy, as well as the http and https proxy listeners.
// Returns a cleanup function for gracefully shutting down the backend
func setupBackend(conf config.Config, httpHandler func(packet.HttpPacket)) (func(), error) {
	logLevel := slog.LevelInfo

	if conf.Debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)

	if err := cacert.InitCaCert(); err != nil {
		return nil, err
	}

	httpListener, err := http.ListenAndServe(conf, httpHandler)

	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	httpsListener, err := http.ListenAndServeTls(conf, httpHandler)

	if err != nil {
		return nil, fmt.Errorf("https server: %w", err)
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
	restart, err := setupBackend(conf, func(p packet.HttpPacket) {
		packetChan <- p
	})

	if err != nil {
		slog.Error("Error setting up backend", "error", err)
		return
	}

	mainWindow := ui.MakeUi(packetChan, func() {
		slog.Info("Restarting backend...")
		restart()
		restart, err = setupBackend(config.ParseFlags(app.Preferences()), func(p packet.HttpPacket) {
			packetChan <- p
		})
	})

	slog.Info("Showing ui...")
	mainWindow.ShowAndRun()
}
