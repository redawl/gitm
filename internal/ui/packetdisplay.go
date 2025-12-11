package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
)

type PacketDisplay struct {
	widget.BaseWidget
	entry       *PacketEntry
	placeHolder *PlaceHolder
	label       *widget.Label

	packet packet.Packet
}

func NewPacketDisplay(title string, w fyne.Window, handleDecodeResult func(string)) *PacketDisplay {
	packetDisplay := &PacketDisplay{
		entry: NewPacketEntry(w, handleDecodeResult),
		label: &widget.Label{
			Text:     title,
			SizeName: theme.SizeNameSubHeadingText,
		},
		placeHolder: NewPlaceHolder(lang.L("Select a packet"), theme.InfoIcon()),
	}

	packetDisplay.ExtendBaseWidget(packetDisplay)

	return packetDisplay
}

func (p *PacketDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewBorder(
			container.NewVBox(p.label, widget.NewSeparator()),
			nil,
			nil,
			nil,
			container.NewStack(
				p.placeHolder,
				p.entry,
			),
		),
	)
}

func (p *PacketDisplay) SetPacket(pack packet.Packet, displayRequest bool) {
	p.packet = pack
	var text string
	if displayRequest {
		text = pack.FormatRequestContent()
	} else {
		text = pack.FormatResponseContent()
	}

	p.placeHolder.Hide()

	p.entry.SetText(text)

	p.entry.ScrollToTop()
}

func (p *PacketDisplay) UnsetPacket() {
	p.placeHolder.Show()
	p.entry.SetText("")
	p.entry.ScrollToTop()
}
