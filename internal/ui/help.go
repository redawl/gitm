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

func MakeHelp(w fyne.Window) *fyne.Menu {
	about := fyne.NewMenuItem(lang.L("About"), func() {
		if about, err := readDocsFile("about.md"); err != nil {
			slog.Error("Error reading about.md")
		} else {
			NewPopoutDialog(lang.L("About"), lang.L("Dismiss"), func() fyne.CanvasObject { return widget.NewRichTextFromMarkdown(about) }, w).Show()
		}
	})

	menu := fyne.NewMenu(lang.L("Help"), MakeDocs(w), about)

	return menu
}

// CreateDocsEntry creates a menu subentry
func CreateDocsEntry(label string, filename string, w fyne.Window) *container.TabItem {
	content := widget.NewRichText()
	if rawContent, err := readDocsFile(filename); err != nil {
		util.ReportUiErrorWithMessage("Error reading docs entry", err, w)
	} else {
		content.ParseMarkdown(rawContent)
	}

	return container.NewTabItem(label, content)
}

func MakeDocs(w fyne.Window) *fyne.MenuItem {
	return fyne.NewMenuItem(lang.L("Documentation"), func() {
		OpenDoc("", w)
	})
}

func OpenDoc(doc string, w fyne.Window) {
	if doc == "" {
		doc = "Home"
	}
	creator := func() fyne.CanvasObject {
		content := container.NewAppTabs(
			CreateDocsEntry(lang.L("Home"), "default.md", w),
			CreateDocsEntry(lang.L("Setup"), "setup.md", w),
			CreateDocsEntry(lang.L("Usage Tips"), "usage.md", w),
			// TODO: How to display editor, since it's designed to be in its own window...
			// container.NewTabItem(lang.L("Docs Editor"), Editor),
		)
		content.SetTabLocation(container.TabLocationLeading)
		for index, item := range content.Items {
			if item.Text == doc {
				content.SelectIndex(index)
				break
			}
		}
		return content
	}

	helpDialog := NewPopoutDialog(lang.L("Documentation"), lang.L("Close"), creator, w)

	helpDialog.Resize(fyne.NewSize(
		w.Canvas().Size().Width*(2.0/3),
		w.Canvas().Size().Height*(2.0/3),
	))

	helpDialog.Show()
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
