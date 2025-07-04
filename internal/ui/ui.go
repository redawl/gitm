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

const RECENTLY_OPENED = "RecentlyOpened"

func makeMenu(packetFilter *PacketFilter, settingsHandler func()) *fyne.MainMenu {
	recentlyOpenedFiles := fyne.CurrentApp().Preferences().StringList(RECENTLY_OPENED)
	recentlyOpenedItem := &fyne.MenuItem{
		Label: "Open Recent",
	}

	recentlyOpenItems := make([]*fyne.MenuItem, len(recentlyOpenedFiles))
	if len(recentlyOpenedFiles) == 0 {
		recentlyOpenedItem.Disabled = true
	} else {
		recentlyOpenedItem.Disabled = false
		for index, recentlyOpened := range recentlyOpenedFiles {
			recentlyOpenItems[index] = fyne.NewMenuItem(recentlyOpened, func() {
				packetFilter.LoadPacketsFromFile(recentlyOpened)
			})
		}
	}

	recentlyOpenedItem.ChildMenu = fyne.NewMenu("", recentlyOpenItems...)
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			&fyne.MenuItem{
				Label:    "Open",
				Action:   packetFilter.LoadPackets,
				Shortcut: OpenShortcut,
			},
			recentlyOpenedItem,
			// TODO: Shortcut
			fyne.NewMenuItem("Clear", packetFilter.ClearPackets),
			&fyne.MenuItem{
				Label:    "Save",
				Action:   packetFilter.SavePackets,
				Shortcut: SaveShortcut,
			},
			&fyne.MenuItem{
				Label:    "Settings",
				Action:   settingsHandler,
				Shortcut: SettingsShortcut,
			},
		),
		MakeHelp(),
	)
	fyne.CurrentApp().Preferences().AddChangeListener(func() {
		recentlyOpenedFiles = fyne.CurrentApp().Preferences().StringList(RECENTLY_OPENED)
		if len(recentlyOpenedFiles) == 0 {
			recentlyOpenedItem.Disabled = true
		} else {
			recentlyOpenedItem.Disabled = false
			newItems := make([]*fyne.MenuItem, len(recentlyOpenedFiles))
			for index, recentlyOpened := range recentlyOpenedFiles {
				newItems[index] = fyne.NewMenuItem(recentlyOpened, func() {
					packetFilter.LoadPacketsFromFile(recentlyOpened)
				})
			}
			recentlyOpenedItem.ChildMenu.Items = newItems
			// TODO: Do I really need to refresh the whole menu?
			mainMenu.Refresh()
		}
	})
	return mainMenu
}

// MakeUi Creates the Fyne UI for GITM, and then runs the UI event loop.
func MakeUi(packetChan chan packet.HttpPacket, restart func()) fyne.Window {
	a := fyne.CurrentApp()

	w := util.NewWindowIfNotExists("Gopher in the middle")
	w.SetMaster()
	w.Resize(fyne.NewSize(
		float32(a.Preferences().FloatWithFallback("MainWindowWidth", 1000)),
		float32(a.Preferences().FloatWithFallback("MainWindowHeight", 800)),
	))

	recordButton := NewRecordButton()

	requestContent := NewPacketDisplay("Request", w)
	responseContent := NewPacketDisplay("Response", w)

	packetFilter := NewPacketFilter(w)

	registerShortcuts(packetFilter, w, restart)

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

	listPlaceHolder := NewPlaceHolder("Record new packets, \nor open a capture file", theme.FolderOpenIcon())

	packetFilter.AddListener(func() {
		if len(packetFilter.FilteredPackets()) > 0 {
			listPlaceHolder.Hide()
		} else {
			listPlaceHolder.Show()
			requestContent.SetText("")
			responseContent.SetText("")
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
			packetFilter,
			func() {
				settings.MakeSettingsUi(restart).Show()
			},
		),
	)

	w.SetContent(masterLayout)

	w.SetOnClosed(func() {
		a.Preferences().SetFloat("MainWindowWidth", float64(w.Canvas().Size().Width))
		a.Preferences().SetFloat("MainWindowHeight", float64(w.Canvas().Size().Height))
	})

	return w
}
