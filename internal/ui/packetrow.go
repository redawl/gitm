package ui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
)

type PacketRow struct {
	widget.BaseWidget
	hostname widget.Label
	request  widget.Label
	response widget.Label
}

type packetRowLayout struct{}

func (pr *packetRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float64(0)

	for _, o := range objects {
		w += o.MinSize().Width
		h = math.Max(float64(o.MinSize().Height), h)
	}

	return fyne.NewSize(w, float32(h))
}

func (pr *packetRowLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	if len(objects) != 3 {
		panic("objects should be length 3!")
	}

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

func (row *PacketRow) UpdateRow(p packet.HttpPacket) {
	path := p.Path

	if len(path) == 0 {
		path = "/"
	}

	if row.hostname.Text != p.Hostname {
		row.hostname.SetText(p.Hostname)
	}

	requestContent := fmt.Sprintf("%s %s %s", p.Method, path, p.ReqProto)

	if row.request.Text != requestContent {
		row.request.SetText(requestContent)
	}

	responseContent := fmt.Sprintf("%s %s", p.RespProto, p.Status)

	if row.response.Text != responseContent {
		row.response.SetText(responseContent)
	}

	row.ExtendBaseWidget(row)
}
