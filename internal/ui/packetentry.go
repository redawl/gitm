package ui

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal"
)

// Right-click menu
var _ fyne.SecondaryTappable = (*PacketEntry)(nil)

// Selecting text
var (
	_ desktop.Mouseable = (*PacketEntry)(nil)
	_ desktop.Hoverable = (*PacketEntry)(nil)
)

// Handle shortcuts
var (
	_ fyne.Shortcutable = (*PacketEntry)(nil)
	_ fyne.Focusable    = (*PacketEntry)(nil)
)

type PacketEntry struct {
	widget.TextGrid
	parent                                                     fyne.Window
	selectStartRow, selectStartCol, selectEndRow, selectEndCol int
	selecting                                                  bool
	handleDecodeResult                                         func(string)
}

func NewPacketEntry(w fyne.Window, handleDecodeResult func(string)) *PacketEntry {
	p := &PacketEntry{
		TextGrid:           widget.TextGrid{Scroll: fyne.ScrollBoth},
		parent:             w,
		handleDecodeResult: handleDecodeResult,
	}

	p.ExtendBaseWidget(p)

	return p
}

func (p *PacketEntry) SetText(text string) {
	colorNormal := &widget.CustomTextGridStyle{
		FGColor: theme.Color(theme.ColorNameForeground),
		BGColor: theme.Color(theme.ColorNameBackground),
	}
	p.selecting = false
	p.SetStyleRange(0, 0, len(p.Rows), len(p.Row(len(p.Rows)-1).Cells), colorNormal)
	p.TextGrid.SetText(text)
	p.ScrollToTop()
}

func (p *PacketEntry) MouseDown(event *desktop.MouseEvent) {
	if event.Button == desktop.MouseButtonPrimary {
		colorNormal := &widget.CustomTextGridStyle{
			FGColor: theme.Color(theme.ColorNameForeground),
			BGColor: theme.Color(theme.ColorNameBackground),
		}

		p.SetStyleRange(0, 0, len(p.Rows), len(p.Row(len(p.Rows)-1).Cells), colorNormal)
		row, col := p.CursorLocationForPosition(event.Position)

		p.selectStartRow = row
		p.selectStartCol = col
		p.selecting = true
		p.Refresh()
	}
}

func (p *PacketEntry) MouseUp(event *desktop.MouseEvent) {
	p.selecting = false
}

func (p *PacketEntry) MouseIn(event *desktop.MouseEvent) {
	p.parent.Canvas().Focus(p)
}

func (p *PacketEntry) MouseMoved(event *desktop.MouseEvent) {
	if p.selecting && event.Button == desktop.MouseButtonPrimary {
		colorHighlighted := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameSelection)}
		colorNormal := &widget.CustomTextGridStyle{
			FGColor: theme.Color(theme.ColorNameForeground),
			BGColor: theme.Color(theme.ColorNameBackground),
		}
		row, col := p.CursorLocationForPosition(event.Position)

		if p.selectEndRow == row && p.selectEndCol == col {
			return
		}

		startRow, startCol, endRow, endCol := p.getActualStartAndEnd()
		p.SetStyleRange(startRow, startCol, endRow, endCol, colorNormal)

		p.selectEndRow = row
		p.selectEndCol = col

		startRow, startCol, endRow, endCol = p.getActualStartAndEnd()

		p.SetStyleRange(startRow, startCol, endRow, endCol, colorHighlighted)
		p.Refresh()
	}
}

func (p *PacketEntry) MouseOut() {
	p.selecting = false
}

func (p *PacketEntry) TypedShortcut(s fyne.Shortcut) {
	switch s.(type) {
	case *fyne.ShortcutCopy:
		p.copyToClipBoard()
	case *fyne.ShortcutSelectAll:
		p.selectAll()
	default:
		slog.Debug("Not handling shortcut", "shortcut", s.ShortcutName())
	}
}

func (p *PacketEntry) FocusGained() {}
func (p *PacketEntry) FocusLost() {
	p.selecting = false
}
func (p *PacketEntry) TypedRune(_ rune)          {}
func (p *PacketEntry) TypedKey(_ *fyne.KeyEvent) {}

func (p *PacketEntry) getActualStartAndEnd() (startRow int, startCol int, endRow int, endCol int) {
	// First normalize end col to make sure they fall within the length of the row
	selectEndCol := min(len(p.Row(p.selectEndRow).Cells)-1, p.selectEndCol)

	if p.selectEndRow == p.selectStartRow {
		if selectEndCol > p.selectStartCol {
			return p.selectStartRow, p.selectStartCol, p.selectEndRow, selectEndCol
		} else {
			return p.selectEndRow, selectEndCol, p.selectStartRow, p.selectStartCol
		}
	} else if p.selectEndRow > p.selectStartRow {
		return p.selectStartRow, p.selectStartCol, p.selectEndRow, selectEndCol
	} else {
		return p.selectEndRow, selectEndCol, p.selectStartRow, p.selectStartCol
	}
}

// SelectedText returns the highlighted text
func (p *PacketEntry) SelectedText() string {
	// Short circuit if -1
	if p.selectStartRow == -1 {
		return ""
	}

	startRow, startCol, endRow, endCol := p.getActualStartAndEnd()

	if startRow == endRow {
		return p.RowText(startRow)[startCol : endCol+1]
	}

	builder := strings.Builder{}
	builder.WriteString(p.RowText(startRow)[startCol:])
	builder.WriteByte('\n')

	if startRow != endRow {
		for i := startRow + 1; i < endRow; i++ {
			builder.WriteString(p.RowText(i))
			builder.WriteByte('\n')
		}

		builder.WriteString(p.RowText(endRow)[:endCol+1])
	}

	return builder.String()
}

// HasSelectedText reports whether there is an user-selected text
func (p *PacketEntry) HasSelectedText() bool {
	valuesToCheck := []int{
		p.selectStartRow,
		p.selectStartCol,
		p.selectEndRow,
		p.selectEndCol,
	}
	for _, value := range valuesToCheck {
		if value > 0 {
			return true
		}
	}
	return false
}

// copyToClipBoard copies the p.SelectedText() to the system clipboard
func (p *PacketEntry) copyToClipBoard() {
	c := fyne.CurrentApp().Clipboard()
	if p.HasSelectedText() {
		selectedText := p.SelectedText()
		c.SetContent(selectedText)
	} else {
		slog.Info("No text selected")
	}
}

func (p *PacketEntry) selectAll() {
	colorHighlighted := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameSelection)}
	p.selectStartRow = 0
	p.selectStartCol = 0
	p.selectEndRow = len(p.Rows) - 1
	p.selectEndCol = len(p.Rows[p.selectEndRow].Cells) - 1

	p.SetStyleRange(p.selectStartRow, p.selectStartCol, p.selectEndRow, p.selectEndCol, colorHighlighted)
	p.Refresh()
}

// TappedSecondary handles when the user right clicks
// Creates the right click menu with entries for the supported decodings
func (p *PacketEntry) TappedSecondary(evt *fyne.PointEvent) {
	c := fyne.CurrentApp().Driver().CanvasForObject(p)

	decodeEntries := make([]*fyne.MenuItem, 0)

	for encodingKey := range GetEncodings() {
		decodeEntries = append(decodeEntries, fyne.NewMenuItem(encodingKey, func() {
			if !p.HasSelectedText() {
				dialog.ShowError(errors.New(lang.L("select text before attempting to decode :)")), p.parent)
				return
			}

			if decoded, err := ExecuteEncoding(encodingKey, p.SelectedText()); err != nil {
				dialog.ShowError(fmt.Errorf("decoding error: %w", err), p.parent)
			} else {
				p.handleDecodeResult(decoded)
			}
		}))
	}

	decodeEntries = append(decodeEntries, fyne.NewMenuItemSeparator())

	customDecodeEntries := fyne.CurrentApp().Preferences().StringList(internal.CustomDecodings)

	for _, decodeEntry := range customDecodeEntries {
		index := strings.Index(decodeEntry, ":")
		if index < 0 {
			slog.Error("Invalid custom decoding", "decoding", decodeEntry)
			continue
		}
		label, command := decodeEntry[0:index], decodeEntry[index+1:]

		decodeEntries = append(decodeEntries, fyne.NewMenuItem(label, func() {
			if !p.HasSelectedText() {
				dialog.ShowError(errors.New(lang.L("select text before attempting to decode :)")), p.parent)
				return
			}

			commandArgs := make([]string, 0, 2)
			if runtime.GOOS != "windows" {
				commandArgs = append(commandArgs, "sh", "-c")
			} else {
				commandArgs = append(commandArgs, "cmd", "/c")
			}
			commandArgs = append(commandArgs, fmt.Sprintf(command, p.SelectedText()))

			cmd := exec.Command(commandArgs[0], commandArgs[1:]...)

			if decoded, err := cmd.Output(); err != nil {
				dialog.ShowError(fmt.Errorf("decoding error: %w", err), p.parent)
			} else {
				p.handleDecodeResult(string(decoded))
			}
		}))
	}

	decodeEntries = append(decodeEntries, fyne.NewMenuItemSeparator())

	copyItem := fyne.NewMenuItem(lang.L("Copy"), p.copyToClipBoard)
	copyItem.Shortcut = &fyne.ShortcutCopy{}

	selectAllItem := fyne.NewMenuItem(lang.L("Select All"), p.selectAll)
	selectAllItem.Shortcut = &fyne.ShortcutSelectAll{}

	decodeEntries = append(decodeEntries, copyItem, selectAllItem)

	menu := fyne.NewMenu(lang.L("Decode"), decodeEntries...)

	popup := widget.NewPopUpMenu(menu, c)
	popup.ShowAtRelativePosition(evt.Position, p)
}
