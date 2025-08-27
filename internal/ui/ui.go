package ui

import (
	"errors"
	"log/slog"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	// packetChan is the communication chan between the backend and frontend.
	// packets come in from the backend, and are processed by the frontend.
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

func (m *MainWindow) updateRecentlyOpenedItems(parent *fyne.MenuItem) {
	recentlyOpenedFiles := fyne.CurrentApp().Preferences().StringList(RECENTLY_OPENED)

	recentlyOpenItems := make([]*fyne.MenuItem, len(recentlyOpenedFiles))
	if len(recentlyOpenedFiles) == 0 {
		parent.Disabled = true
	} else {
		parent.Disabled = false
		for index, recentlyOpened := range recentlyOpenedFiles {
			recentlyOpenItems[index] = fyne.NewMenuItem(recentlyOpened, func() {
				m.PacketFilter.LoadPacketsFromFile(recentlyOpened)
			})
			if _, err := os.Stat(recentlyOpened); errors.Is(err, os.ErrNotExist) {
				recentlyOpenItems[index].Disabled = true
				recentlyOpenItems[index].Icon = theme.BrokenImageIcon()
			} else {
				recentlyOpenItems[index].Icon = theme.FileApplicationIcon()
			}
		}
		if parent.ChildMenu == nil {
			parent.ChildMenu = fyne.NewMenu("")
		}
		parent.ChildMenu.Items = recentlyOpenItems
	}
	// TODO: Do I really need to refresh the whole menu?
	m.MainMenu().Refresh()
}

// makeMenu creates the main menu for the master GITM window
func (m *MainWindow) makeMenu(settingsHandler func()) {
	recentlyOpenedItem := &fyne.MenuItem{
		Label: lang.L("Open Recent"),
	}

	m.updateRecentlyOpenedItems(recentlyOpenedItem)

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu(lang.L("File"),
			&fyne.MenuItem{Label: lang.L("Open"), Action: m.PacketFilter.LoadPackets, Shortcut: OpenShortcut},
			recentlyOpenedItem,
			&fyne.MenuItem{Label: lang.L("Clear"), Action: m.PacketFilter.ClearPackets, Shortcut: ClearShortcut},
			&fyne.MenuItem{Label: lang.L("Save"), Action: m.PacketFilter.SavePackets, Shortcut: SaveShortcut},
			&fyne.MenuItem{Label: lang.L("Settings"), Action: settingsHandler, Shortcut: SettingsShortcut},
			fyne.NewMenuItemSeparator(),
			&fyne.MenuItem{Label: lang.L("Quit"), Action: fyne.CurrentApp().Quit, Shortcut: QuitShortcut, IsQuit: true},
		),
		MakeHelp(m),
	)
	fyne.CurrentApp().Preferences().AddChangeListener(func() {
		m.updateRecentlyOpenedItems(recentlyOpenedItem)
	})
	m.SetMainMenu(mainMenu)
}

// MakeMainWindow creates the Fyne UI for GITM
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

	mainWindow.makeMenu(func() { settings.MakeSettingsUi(w, restart).Show() })

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

	mainWindow.CheckForCrashData()

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

// CheckForCrashData checks to see if there is data from a prior crash.
//
// If there is, it displays a comfirmation dialog for the user to choose whether to load the
// crash data.
func (m *MainWindow) CheckForCrashData() {
	configDir, err := util.GetConfigDir()
	if err != nil {
		slog.Error("Error getting config location", "error", err)
	}
	crashData := configDir + string(os.PathSeparator) + "crash.json"
	if _, err := os.Stat(crashData); err == nil {
		dialog.NewConfirm(lang.L("Crash data found"), lang.L("There was crash data found. Want to load it?"), func(b bool) {
			if b {
				m.PacketFilter.LoadPacketsFromFile(crashData)
			}
			if err := os.Remove(crashData); err != nil {
				util.ReportUiError(err, m)
			}
		}, m).Show()
	}
}
