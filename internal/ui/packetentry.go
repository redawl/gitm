package ui

import (
	"log/slog"

	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PacketEntry struct {
    widget.TextGrid
    selectStartRow, selectStartCol, selectEndRow, selectEndCol int
    selecting bool
}

func (p *PacketEntry) MouseDown(event *desktop.MouseEvent) {
    TEXTGRID_COLOR_NORMAL := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameBackground)}

    p.SetStyleRange(0, 0, len(p.Rows), len(p.Row(len(p.Rows)-1).Cells), TEXTGRID_COLOR_NORMAL)
    row, col := p.CursorLocationForPosition(event.Position)

    p.selectStartRow = row
    p.selectStartCol = col
    p.selecting = true
    p.Refresh()
}

func (p *PacketEntry) MouseUp(event *desktop.MouseEvent) {
    p.selecting = false
}

func (p *PacketEntry) MouseIn(event *desktop.MouseEvent) {}

func (p *PacketEntry) MouseMoved(event *desktop.MouseEvent) {
    if p.selecting {
        TEXTGRID_COLOR_HIGHLIGHTED := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameSelection)}
        TEXTGRID_COLOR_NORMAL := &widget.CustomTextGridStyle{BGColor: theme.Color(theme.ColorNameBackground)}
        row, col := p.CursorLocationForPosition(event.Position)
        startRow, startCol, endRow, endCol := p.getActualStartAndEnd()
        p.SetStyleRange(startRow, startCol, endRow, endCol, TEXTGRID_COLOR_NORMAL)

        p.selectEndRow = row
        p.selectEndCol = col

        startRow, startCol, endRow, endCol = p.getActualStartAndEnd()

        p.SetStyleRange(startRow, startCol, endRow, endCol, TEXTGRID_COLOR_HIGHLIGHTED)
        p.Refresh()
    }
}

func (p *PacketEntry) MouseOut() {}

func (p *PacketEntry) getActualStartAndEnd() (startRow int, startCol int, endRow int, endCol int) {
    // First normalize end col to make sure they fall within the length of the row
    selectEndCol := min(len(p.Row(p.selectEndRow).Cells) - 1, p.selectEndCol)

    slog.Info("", "selectEndCol", selectEndCol)
    slog.Info("", "len(p.Row(p.selectEndRow).Cells)", len(p.Row(p.selectEndRow).Cells), "p.selectEndCol", p.selectEndCol)

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

