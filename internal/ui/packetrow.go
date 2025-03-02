package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
)

type PacketRow struct {
    widget.Label
}

func NewPacketRow() *PacketRow {
    row := &PacketRow{
        Label: widget.Label{
            TextStyle: fyne.TextStyle{
                Monospace: true,
            },
        },
    }

    row.ExtendBaseWidget(row)
    return row
}

func (row *PacketRow) UpdateRow (p packet.HttpPacket) {
    path := p.Path

    if len(path) == 0 {
        path = "/"
    } else if len(path) > 100 {
        path = path[:100] + "..."
    }

    row.SetText(fmt.Sprintf("%s: %s %s %s -> %s %s", p.Hostname, p.Method, path, p.ReqProto, p.RespProto, p.Status))
    row.Refresh()
}

