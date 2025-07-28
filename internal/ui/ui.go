package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/ui/settings"
	"github.com/redawl/gitm/internal/util"
)

const RECENTLY_OPENED = "RecentlyOpened"

// MainWindow is the main window of GITM
// It is the window that opens when the application is first launched.
type MainWindow struct {
	fyne.Window
	// packetChan
	// TODO: docs
	packetChan chan packet.Packet
	// requestContent the request content of the currently selected packet
	requestContent *PacketDisplay
	// responseContent the response content of the currently selected packet
	responseContent *PacketDisplay
	// recordButton is the button that starts/stops recording when pressed
	recordButton *RecordButton
	// PacketFilter manages the list of packets currently loaded into GITM
	PacketFilter *PacketFilter
}

func makeMenu(packetFilter *PacketFilter, settingsHandler func()) *fyne.MainMenu {
	recentlyOpenedFiles := fyne.CurrentApp().Preferences().StringList(RECENTLY_OPENED)
	recentlyOpenedItem := &fyne.MenuItem{
		Label: lang.L("Open Recent"),
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
			recentlyOpenItems[index].Icon = theme.FileApplicationIcon()
		}
	}

	recentlyOpenedItem.ChildMenu = fyne.NewMenu("", recentlyOpenItems...)
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu(lang.L("File"),
			&fyne.MenuItem{Label: lang.L("Open"), Action: packetFilter.LoadPackets, Shortcut: OpenShortcut},
			recentlyOpenedItem,
			&fyne.MenuItem{Label: lang.L("Clear"), Action: packetFilter.ClearPackets, Shortcut: ClearShortcut},
			&fyne.MenuItem{Label: lang.L("Save"), Action: packetFilter.SavePackets, Shortcut: SaveShortcut},
			&fyne.MenuItem{Label: lang.L("Settings"), Action: settingsHandler, Shortcut: SettingsShortcut},
			fyne.NewMenuItemSeparator(),
			&fyne.MenuItem{Label: lang.L("Quit"), Action: fyne.CurrentApp().Quit, Shortcut: QuitShortcut, IsQuit: true},
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

// MakeMainWindow Creates the Fyne UI for GITM
func MakeMainWindow(packetChan chan packet.Packet, restart func()) *MainWindow {
	w := util.NewWindowIfNotExists(lang.L("GITM"))
	filter := NewPacketFilter(w)
	w.SetMaster()
	mainWindow := &MainWindow{
		Window:          w,
		requestContent:  NewPacketDisplay(lang.L("Request"), w),
		responseContent: NewPacketDisplay(lang.L("Response"), w),
		recordButton:    NewRecordButton(filter, w),
		PacketFilter:    filter,
		packetChan:      packetChan,
	}
	mainWindow.registerShortcuts(restart)

	mainWindow.SetMainMenu(
		makeMenu(
			mainWindow.PacketFilter,
			func() { settings.MakeSettingsUi(restart).Show() },
		),
	)

	mainWindow.SetContent(
		container.NewVSplit(
			container.NewBorder(
				container.NewVBox(
					mainWindow.recordButton,
					mainWindow.PacketFilter,
					widget.NewSeparator(),
				),
				nil,
				nil,
				nil,
				NewPacketList(mainWindow.PacketFilter, mainWindow),
			),
			container.NewHSplit(
				mainWindow.requestContent,
				mainWindow.responseContent,
			),
		),
	)

	mainWindow.StartPacketHandler()

	return mainWindow
}

func (m *MainWindow) StartPacketHandler() {
	go func() {
		for {
			p := <-m.packetChan
			if m.recordButton.IsRecording {
				existingPacket := m.PacketFilter.FindPacket(p)

				if existingPacket != nil {
					existingPacket.UpdatePacket(p)
				} else {
					m.PacketFilter.AppendPacket(p)
				}
			}
		}
	}()
}
