package main

import (
	"fmt"
	"log/slog"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/storage/repository"
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

	if conf.EnablePacServer {
		go func() {
			if err := socks5.ListenAndServePac(&conf); err != nil {
				slog.Error("Error starting pac server", "error", err)
			}
		}()
	}

	return func() { _ = socksListener.Close }, nil
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

	slog.Info("Showing ui...")
	mainWindow := ui.MakeMainWindow(packetChan, func() {
		slog.Info("Restarting backend...")
		restart()
		restart, err = setupBackend(config.FromPreferences(app.Preferences()), func(p packet.Packet) {
			packetChan <- p
		})
	})
	mainWindow.ShowAndRun()
}
