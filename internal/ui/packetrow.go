package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/util"
)

type PacketRow struct {
	widget.BaseWidget
	hostname widget.Label
	request  widget.Label
	response widget.Label
}

type packetRowLayout struct{}

func (pr *packetRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)

	for _, o := range objects {
		w += o.MinSize().Width
		if o.MinSize().Height > h {
			h = o.MinSize().Height
		}
	}

	return fyne.NewSize(w, h)
}

func (pr *packetRowLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	util.Assert(len(objects) == 3)

	commonHeight := containerSize.Height - pr.MinSize(objects).Height

	hostname, request, response := objects[0], objects[1], objects[2]

	hostname.Resize(fyne.NewSize(containerSize.Width*.25, hostname.MinSize().Height))
	hostname.Move(fyne.NewPos(0, commonHeight))

	request.Resize(fyne.NewSize(containerSize.Width*.60, request.MinSize().Height))
	request.Move(fyne.NewPos(containerSize.Width*.25, commonHeight))

	response.Resize(fyne.NewSize(containerSize.Width*.15, response.MinSize().Height))
	response.Move(fyne.NewPos(containerSize.Width*.85, commonHeight))
}

func NewPacketRow() *PacketRow {
	row := &PacketRow{
		hostname: widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Truncation: fyne.TextTruncateEllipsis,
		},
		request: widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Truncation: fyne.TextTruncateEllipsis,
		},
		response: widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Truncation: fyne.TextTruncateEllipsis,
		},
	}

	row.ExtendBaseWidget(row)
	return row
}

func (row *PacketRow) CreateRenderer() fyne.WidgetRenderer {
	c := container.New(&packetRowLayout{}, &row.hostname, &row.request, &row.response)

	return widget.NewSimpleRenderer(c)
}

func (row *PacketRow) UpdateRow(p packet.Packet) {
	hostname := p.FormatHostname()
	if row.hostname.Text != hostname {
		row.hostname.SetText(hostname)
	}

	requestLine := p.FormatRequestLine()

	if row.request.Text != requestLine {
		row.request.SetText(requestLine)
	}

	responseLine := p.FormatResponseLine()

	if row.response.Text != responseLine {
		if httpPacket, ok := p.(*packet.HttpPacket); ok && len(httpPacket.Status) > 0 {
			switch strings.Split(httpPacket.Status, " ")[0][0] {
			case '2':
				row.response.Importance = widget.SuccessImportance
			case '3':
				row.response.Importance = widget.HighImportance
			case '4':
				row.response.Importance = widget.DangerImportance
			case '5':
				row.response.Importance = widget.WarningImportance
			}
		}
		row.response.SetText(responseLine)
	}

	row.ExtendBaseWidget(row)
}
