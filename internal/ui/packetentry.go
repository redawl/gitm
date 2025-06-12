package ui

import (
	"fmt"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// For rightclick menu
var _ fyne.SecondaryTappable = (*PacketEntry)(nil)

// For selecting text
var _ desktop.Mouseable = (*PacketEntry)(nil)
var _ desktop.Hoverable = (*PacketEntry)(nil)

type PacketEntry struct {
	widget.TextGrid
	selectStartRow, selectStartCol, selectEndRow, selectEndCol int
	selecting                                                  bool
}

func NewPacketEntry() *PacketEntry {
	p := &PacketEntry{
		TextGrid: widget.TextGrid{Scroll: fyne.ScrollBoth},
	}

	p.ExtendBaseWidget(p)

	return p
}

func (p *PacketEntry) SetText(text string) {
	TEXTGRID_COLOR_NORMAL := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameBackground)}
	p.selecting = false

	p.SetStyleRange(0, 0, len(p.Rows), len(p.Row(len(p.Rows)-1).Cells), TEXTGRID_COLOR_NORMAL)
	p.TextGrid.SetText(text)
}

func (p *PacketEntry) MouseDown(event *desktop.MouseEvent) {
	if event.Button == desktop.MouseButtonPrimary {
		TEXTGRID_COLOR_NORMAL := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameBackground)}

		p.SetStyleRange(0, 0, len(p.Rows), len(p.Row(len(p.Rows)-1).Cells), TEXTGRID_COLOR_NORMAL)
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
	if event.Button != desktop.MouseButtonPrimary {
		p.selecting = false
	}
}

func (p *PacketEntry) MouseMoved(event *desktop.MouseEvent) {
	if p.selecting && event.Button == desktop.MouseButtonPrimary {
		TEXTGRID_COLOR_HIGHLIGHTED := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameSelection)}
		TEXTGRID_COLOR_NORMAL := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameBackground)}
		row, col := p.CursorLocationForPosition(event.Position)

		if p.selectEndRow == row && p.selectEndCol == col {
			return
		}

		startRow, startCol, endRow, endCol := p.getActualStartAndEnd()
		p.SetStyleRange(startRow, startCol, endRow, endCol, TEXTGRID_COLOR_NORMAL)

		p.selectEndRow = row
		p.selectEndCol = col

		startRow, startCol, endRow, endCol = p.getActualStartAndEnd()

		p.SetStyleRange(startRow, startCol, endRow, endCol, TEXTGRID_COLOR_HIGHLIGHTED)
		p.Refresh()
	}
}

func (p *PacketEntry) MouseOut() {
	p.selecting = false
}

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
	return p.SelectedText() != ""
}

// TappedSecondary handle when the user right clicks
// Creates the right click menu with entries for the supported decodings
func (p *PacketEntry) TappedSecondary(evt *fyne.PointEvent) {
	c := fyne.CurrentApp().Driver().CanvasForObject(p)

	// TODO: Is there a better way to do this?
	if len(fyne.CurrentApp().Driver().AllWindows()) == 0 {
		slog.Error("Failed to create right-click menu, there are no windows ???")
		return
	}

	w := fyne.CurrentApp().Driver().AllWindows()[0]
	decodeEntries := make([]*fyne.MenuItem, 0)

	for encodingKey := range GetEncodings() {
		decodeEntries = append(decodeEntries, fyne.NewMenuItem(encodingKey, func() {
			if !p.HasSelectedText() {
				dialog.NewError(fmt.Errorf("Select text before attempting to decode :)"), w).Show()
				return
			}

			if decoded, err := ExecuteEncoding(encodingKey, p.SelectedText()); err != nil {
				dialog.NewError(fmt.Errorf("Decoding error: %w", err), w).Show()
			} else {
				dialog.NewInformation("Decode result", string(decoded), w).Show()
			}
		}))
	}

	menu := fyne.NewMenu("Decode", decodeEntries...)

	popup := widget.NewPopUpMenu(menu, c)
	popup.ShowAtPosition(evt.AbsolutePosition)
}
