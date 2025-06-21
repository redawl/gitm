package ui

import (
	_ "embed"
	"fmt"
	"io"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

//go:embed docs/about.md
var about string

func MakeHelp() *fyne.Menu {
	app := fyne.CurrentApp()
	about := fyne.NewMenuItem("About", func() {
		w := app.NewWindow("About")

		w.SetContent(widget.NewRichTextFromMarkdown(about))
		w.Show()
	})

	menu := fyne.NewMenu("Help", MakeDocs(), about)

	return menu
}

func MakeDocs() *fyne.MenuItem {
	return fyne.NewMenuItem("Documentation", func() {
		w := fyne.CurrentApp().NewWindow("Documentation")

		content := widget.NewRichText()
		if homeContent, err := readDocsFile("default.md"); err != nil {
			slog.Error("Error reading default", "error", err)
		} else {
			content.ParseMarkdown(homeContent)
		}

		menu := widget.NewMenu(fyne.NewMenu("",
			fyne.NewMenuItem("Home", func() {
				if homeContent, err := readDocsFile("default.md"); err != nil {
					slog.Error("Error reading default", "error", err)
				} else {
					content.ParseMarkdown(homeContent)
				}
			}),
			fyne.NewMenuItem("Iphone setup", func() {
				if iphoneSetupContent, err := readDocsFile("setup-iphone.md"); err != nil {
					slog.Error("Error reading Iphone setup", "error", err)
				} else {
					content.ParseMarkdown(iphoneSetupContent)
				}
			}),
		))

		w.SetContent(container.NewBorder(nil, nil, menu, nil, container.NewVScroll(content)))
		w.Show()
	})
}

// readDocsFile reads filename from the embed storage for docs files
func readDocsFile(filename string) (string, error) {
	uri, err := storage.ParseURI(fmt.Sprintf("docs:%s", filename))
	if err != nil {
		return "", fmt.Errorf("parsing uri: %w", err)
	}
	reader, err := storage.Reader(uri)
	if err != nil {
		return "", fmt.Errorf("init reader: %w", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("reading contents: %w", err)
	}

	return string(content), nil
}
