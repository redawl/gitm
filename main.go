package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/storage/repository"
	"fyne.io/fyne/v2/theme"
	"github.com/redawl/gitm/docs"
	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/socks5"
	"github.com/redawl/gitm/internal/ui"
	"github.com/redawl/gitm/internal/ui/settings"
)

// setupBackend sets up the socks5 proxy.
// Returns a cleanup function for gracefully shutting down the backend
func setupBackend(conf config.Config, httpHandler func(packet.Packet)) (func(), error) {
	logLevel := slog.LevelInfo

	if conf.Debug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)

	socksListener, err := socks5.ListenAndServeSocks5(conf, httpHandler)
	if err != nil {
		return nil, fmt.Errorf("socks5 proxy: %w", err)
	}

	var server *http.Server
	if conf.EnablePacServer {
		server = socks5.SetupPac(&conf)
		go func() {
			if err := server.ListenAndServe(); err != nil {
				slog.Error("Error serving pack", "error", err)
			}
		}()
	}

	return func() {
		if err := socksListener.Close(); err != nil {
			slog.Error("Error closing socks listener", "error", err)
		}
		if err := server.Close(); err != nil {
			slog.Error("Error closing pac server", "error", err)
		}
	}, nil
}

func main() {
	app := app.NewWithID("com.github.redawl.gitm")
	conf := config.FromPreferences(app.Preferences())

	packetChan := make(chan packet.Packet)

	slog.Info("Starting backend...")
	restart, err := setupBackend(conf, func(p packet.Packet) {
		packetChan <- p
	})
	if err != nil {
		slog.Error("Error setting up backend", "error", err)
		settings.MakeSettingsUi(nil).ShowAndRun()

		return
	}

	repository.Register("docs", &docs.DocsRepository{})

	if conf.Theme == "" {
	} else if reader, err := os.Open(conf.Theme); err != nil {
		slog.Error("Error opening theme, falling back to default", "error", err)
		app.Settings().SetTheme(nil)
	} else if th, err := theme.FromJSONReader(reader); err != nil {
		slog.Error("Error loading theme, falling back to default", "error", err)
		app.Settings().SetTheme(nil)
	} else {
		app.Settings().SetTheme(th)
	}

	slog.Info("Showing ui...")
	mainWindow := ui.MakeMainWindow(packetChan, func() {
		slog.Info("Restarting backend...")
		restart()
		restart, err = setupBackend(config.FromPreferences(app.Preferences()), func(p packet.Packet) {
			packetChan <- p
		})
		if err != nil || restart == nil {
			slog.Error("Error setting up backend", "error", err)
			restart = func() {}
		}
	})
	mainWindow.ShowAndRun()
}
