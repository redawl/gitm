package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/ui/settings"
	"github.com/redawl/gitm/internal/util"
)

func makeMenu(clearHandler func(), saveHandler func(), loadHandler func(), settingsHandler func()) *fyne.MainMenu {
	saveItem := fyne.NewMenuItem("Save", saveHandler)
	saveItem.Shortcut = SaveShortcut
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Load", loadHandler),
			fyne.NewMenuItem("Clear", clearHandler),
			saveItem,
			fyne.NewMenuItem("Settings", settingsHandler),
		),
		MakeHelp(),
	)
	return mainMenu
}

// MakeUi Creates the Fyne UI for GITM, and then runs the UI event loop.
func MakeUi(packetChan chan packet.HttpPacket, restart func()) fyne.Window {
	a := fyne.CurrentApp()

	recordButton := NewRecordButton()

	w := a.NewWindow("Gopher in the middle")
	w.SetMaster()

	// TODO: Remove hardcoded default size.
	// We'll need to figure out how we want the application to look when
	// first opened. It would be nice to simulate a "Windowed fullscreen/borderless" look,
	// but fyne does not have direct support.
	w.Resize(fyne.NewSize(1920, 1080))

	requestContent := NewPacketDisplay("Request", w)
	responseContent := NewPacketDisplay("Response", w)

	packetFilter := NewPacketFilter(w)

	registerShortcuts(packetFilter, w)

	uiList := widget.NewList(func() int {
		return len(packetFilter.FilteredPackets())
	}, func() fyne.CanvasObject {
		return NewPacketRow()
	}, func(li widget.ListItemID, co fyne.CanvasObject) {
		row := co.(*PacketRow)
		filteredPackets := packetFilter.FilteredPackets()
		if li < len(filteredPackets) && filteredPackets[li] != nil {
			p := filteredPackets[li]
			row.UpdateRow(*p)
		}
	})
	uiList.HideSeparators = true

	packetFilter.AddListener(uiList.Refresh)

	listPlaceHolder := NewPlaceHolder("Record new packets or open a capture file", theme.FolderOpenIcon())

	packetFilter.AddListener(func() {
		if len(packetFilter.FilteredPackets()) > 0 {
			listPlaceHolder.Hide()
		} else {
			listPlaceHolder.Show()
		}
	})

	uiList.OnSelected = func(id widget.ListItemID) {
		filteredPackets := packetFilter.FilteredPackets()
		util.Assert(id < len(filteredPackets))

		requestContent.SetText(filteredPackets[id].FormatRequestContent())
		responseContent.SetText(filteredPackets[id].FormatResponseContent())
	}

	go func() {
		for {
			p := <-packetChan
			if recordButton.IsRecording {
				existingPacket := packetFilter.FindPacket(&p)

				if existingPacket != nil {
					existingPacket.UpdatePacket(&p)
				} else {
					packetFilter.AppendPacket(&p)
				}

				fyne.Do(uiList.Refresh)
			}
		}
	}()

	packetListContainer := container.NewBorder(
		container.NewVBox(
			container.NewHBox(recordButton),
			container.NewBorder(
				nil,
				nil,
				widget.NewLabel("Filter packets"),
				nil,
				packetFilter,
			),
		),
		nil,
		nil,
		nil,
		container.NewStack(
			container.NewCenter(listPlaceHolder),
			uiList,
		),
	)

	masterLayout := container.NewVSplit(packetListContainer,
		container.NewHSplit(
			requestContent,
			responseContent,
		),
	)

	w.SetMainMenu(
		makeMenu(
			packetFilter.ClearPackets,
			packetFilter.SavePackets,
			packetFilter.LoadPackets,
			func() {
				settings.MakeSettingsUi(restart).Show()
			},
		),
	)

	w.SetContent(masterLayout)

	return w
}
