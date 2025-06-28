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
	"github.com/redawl/gitm/internal/util"
)

//go:embed docs/about.md
var about string

func MakeHelp() *fyne.Menu {
	app := fyne.CurrentApp()
	about := fyne.NewMenuItem("About", func() {
		for _, window := range app.Driver().AllWindows() {
			if window.Title() == "About" {
				window.RequestFocus()
				return
			}
		}
		w := app.NewWindow("About")

		w.SetContent(widget.NewRichTextFromMarkdown(about))
		w.Show()
	})

	menu := fyne.NewMenu("Help", MakeDocs(), about)

	return menu
}

// CreateDocsEntry creates a menu subentry
func CreateDocsEntry(label string, filename string, contentContainer *container.Scroll) *fyne.MenuItem {
	content, ok := contentContainer.Content.(*widget.RichText)
	util.Assert(ok)

	return fyne.NewMenuItem(label, func() {
		if homeContent, err := readDocsFile(filename); err != nil {
			slog.Error("Error reading docs entry", "filename", filename, "error", err)
		} else {
			content.ParseMarkdown(homeContent)
			contentContainer.ScrollToTop()
		}
	})
}

func MakeDocs() *fyne.MenuItem {
	return fyne.NewMenuItem("Documentation", func() {
		w := util.NewWindowIfNotExists("Documentation")

		content := widget.NewRichText()
		contentContainer := container.NewVScroll(content)
		if homeContent, err := readDocsFile("default.md"); err != nil {
			slog.Error("Error reading default", "error", err)
		} else {
			content.ParseMarkdown(homeContent)
		}

		// TODO: Replace with list widget, so the currently selected
		// menu item can be highlighted
		menu := widget.NewMenu(fyne.NewMenu("",
			CreateDocsEntry("Home", "default.md", contentContainer),
			CreateDocsEntry("Iphone setup", "setup-iphone.md", contentContainer),
			CreateDocsEntry("Firefox setup", "firefox.md", contentContainer),
		))

		w.SetContent(container.NewBorder(nil, nil, menu, nil, contentContainer))
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
