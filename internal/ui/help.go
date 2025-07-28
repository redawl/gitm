package ui

import (
	"fmt"
	"io"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/util"
)

func MakeHelp() *fyne.Menu {
	about := fyne.NewMenuItem(lang.L("About"), func() {
		if about, err := readDocsFile("about.md"); err != nil {
			slog.Error("Error reading about.md")
		} else {
			w := util.NewWindowIfNotExists("About")
			w.SetContent(widget.NewRichTextFromMarkdown(about))
			w.Show()
		}
	})

	menu := fyne.NewMenu(lang.L("Help"), MakeDocs(), about)

	return menu
}

// CreateDocsEntry creates a menu subentry
func CreateDocsEntry(label string, filename string, contentContainer *container.Scroll, w fyne.Window) *fyne.MenuItem {
	content, ok := contentContainer.Content.(*widget.RichText)
	util.Assert(ok)

	return fyne.NewMenuItem(label, func() {
		if rawContent, err := readDocsFile(filename); err != nil {
			util.ReportUiErrorWithMessage("Error reading docs entry", err, w)
		} else {
			content.ParseMarkdown(rawContent)
			contentContainer.ScrollToTop()
		}
	})
}

func MakeDocs() *fyne.MenuItem {
	return fyne.NewMenuItem(lang.L("Documentation"), func() {
		w := util.NewWindowIfNotExists(lang.L("Documentation"))

		content := widget.NewRichText()
		content.Wrapping = fyne.TextWrapWord
		contentContainer := container.NewVScroll(content)
		if homeContent, err := readDocsFile("default.md"); err != nil {
			util.ReportUiErrorWithMessage("Error reading default", err, w)
		} else {
			content.ParseMarkdown(homeContent)
		}

		// TODO: Replace with list widget, so the currently selected
		// menu item can be highlighted
		menu := widget.NewMenu(fyne.NewMenu("",
			CreateDocsEntry(lang.L("Home"), "default.md", contentContainer, w),
			CreateDocsEntry(lang.L("Setup"), "setup.md", contentContainer, w),
			CreateDocsEntry(lang.L("Usage Tips"), "usage.md", contentContainer, w),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(lang.L("Docs Editor"), Editor),
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
	w := util.NewWindowIfNotExists(lang.L("Editor"))

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
		fyne.NewMenu(lang.L("File"),
			fyne.NewMenuItem(lang.L("Open"), func() {
				dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						util.ReportUiErrorWithMessage("Error opening file", err, w)
						return
					}

					if reader == nil {
						return
					}

					contents, err := io.ReadAll(reader)
					if err != nil {
						util.ReportUiErrorWithMessage("Error reading file contents", err, w)
						return
					}

					entry.SetText(string(contents))
				}, w).Show()
			}),
			fyne.NewMenuItem(lang.L("Save"), func() {
				dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						util.ReportUiErrorWithMessage("Error saving to file", err, w)
						return
					}

					if writer == nil {
						return
					}

					if _, err := writer.Write([]byte(entry.Text)); err != nil {
						util.ReportUiErrorWithMessage("Error writing file contents", err, w)
					}
				}, w).Show()
			}),
		),
	))
	w.Show()
}
