package ui

import (
	_ "embed"
	"fmt"
	"io"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/util"
)

func MakeHelp() *fyne.Menu {
	about := fyne.NewMenuItem("About", func() {
		if about, err := readDocsFile("about.md"); err != nil {
			slog.Error("Error reading about.md")
		} else {
			w := util.NewWindowIfNotExists("About")
			w.SetContent(widget.NewRichTextFromMarkdown(about))
			w.Show()
		}
	})

	menu := fyne.NewMenu("Help", MakeDocs(), about)

	return menu
}

// CreateDocsEntry creates a menu subentry
func CreateDocsEntry(label string, filename string, contentContainer *container.Scroll) *fyne.MenuItem {
	content, ok := contentContainer.Content.(*widget.RichText)
	util.Assert(ok)

	return fyne.NewMenuItem(label, func() {
		if rawContent, err := readDocsFile(filename); err != nil {
			slog.Error("Error reading docs entry", "filename", filename, "error", err)
		} else {
			content.ParseMarkdown(rawContent)
			contentContainer.ScrollToTop()
		}
	})
}

func MakeDocs() *fyne.MenuItem {
	return fyne.NewMenuItem("Documentation", func() {
		w := util.NewWindowIfNotExists("Documentation")

		content := widget.NewRichText()
		content.Wrapping = fyne.TextWrapWord
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
			CreateDocsEntry("Setup", "setup.md", contentContainer),
			fyne.NewMenuItem("Docs Editor", Editor),
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

func Editor() {
	w := util.NewWindowIfNotExists("Editor")

	entry := widget.NewMultiLineEntry()
	entry.Scroll = fyne.ScrollBoth
	entry.TextStyle = fyne.TextStyle{
		Monospace: true,
	}
	display := widget.NewRichText()
	display.Scroll = fyne.ScrollBoth

	entry.OnChanged = func(s string) {
		display.ParseMarkdown(s)
	}
	w.SetContent(container.NewHSplit(entry, display))
	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open", func() {
				dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						slog.Error("Error opening file", "error", err)
						return
					}

					if reader == nil {
						return
					}

					contents, err := io.ReadAll(reader)
					if err != nil {
						slog.Error("Error reading file contents", "error", err)
						return
					}

					entry.SetText(string(contents))
				}, w).Show()
			}),
			fyne.NewMenuItem("Save", func() {
				dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						slog.Error("Error saving to file", "error", err)
						return
					}

					if writer == nil {
						return
					}

					if _, err := writer.Write([]byte(entry.Text)); err != nil {
						slog.Error("Error writing file contents", "error", err)
					}
				}, w).Show()
			}),
		),
	))
	w.Show()
}
