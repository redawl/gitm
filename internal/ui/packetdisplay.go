package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PacketDisplay struct {
	widget.BaseWidget
	entry       *PacketEntry
	placeHolder *PlaceHolder
	label       *widget.Label
}

func NewPacketDisplay(label string, w fyne.Window) *PacketDisplay {
	packetDisplay := &PacketDisplay{
		entry: NewPacketEntry(w),
		label: &widget.Label{
			Text:     label,
			SizeName: theme.SizeNameSubHeadingText,
		},
		placeHolder: NewPlaceHolder("Select a packet", theme.InfoIcon()),
	}

	packetDisplay.ExtendBaseWidget(packetDisplay)

	return packetDisplay
}

func (pd *PacketDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewBorder(
			pd.label,
			nil,
			nil,
			nil,
			container.NewStack(
				pd.placeHolder,
				pd.entry,
			),
		),
	)
}

func (pd *PacketDisplay) SetText(text string) {
	if len(text) > 0 {
		pd.placeHolder.Hide()
	} else {
		pd.placeHolder.Show()
	}

	pd.entry.SetText(text)

	pd.entry.ScrollToTop()
}
