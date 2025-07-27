package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/util"
)

type PacketList struct {
	widget.BaseWidget
	list        *widget.List
	placeholder *PlaceHolder
}

func NewPacketList(packetFilter *PacketFilter, mainWindow *MainWindow) *PacketList {
	newList := &PacketList{
		list: &widget.List{
			Length:     func() int { return len(packetFilter.FilteredPackets()) },
			CreateItem: func() fyne.CanvasObject { return NewPacketRow() },
			UpdateItem: func(id widget.ListItemID, item fyne.CanvasObject) {
				row := item.(*PacketRow)
				filteredPackets := packetFilter.FilteredPackets()
				if id < len(filteredPackets) && filteredPackets[id] != nil {
					p := filteredPackets[id]
					row.UpdateRow(p)
				}
			},
			OnSelected: func(id widget.ListItemID) {
				filteredPackets := packetFilter.FilteredPackets()
				util.Assert(id < len(filteredPackets))
				p := filteredPackets[id]

				mainWindow.requestContent.SetText(p.FormatRequestContent())
				mainWindow.responseContent.SetText(p.FormatResponseContent())
			},
			HideSeparators: true,
		},
		placeholder: NewPlaceHolder(lang.L("Record new packets, \nor open a capture file"), theme.FolderOpenIcon()),
	}

	packetFilter.AddListener(func() {
		if len(packetFilter.FilteredPackets()) > 0 {
			newList.placeholder.Hide()
		} else {
			newList.placeholder.Show()
			mainWindow.requestContent.SetText("")
			mainWindow.responseContent.SetText("")
		}
	})

	packetFilter.AddListener(newList.list.Refresh)

	newList.ExtendBaseWidget(newList)

	return newList
}

func (p *PacketList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewStack(
			p.placeholder,
			p.list,
		),
	)
}

func (p *PacketList) MinSize() fyne.Size {
	return p.list.MinSize()
}
