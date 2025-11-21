package ui

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

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

const RecentlyOpened = "RecentlyOpened"

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
	// analysisToolbar is the top-level toolbar
	analysisToolbar *AnalysisToolbar
	// PacketFilter manages the list of packets currently loaded into GITM
	PacketFilter *PacketFilter
}

func (m *MainWindow) updateRecentlyOpenedItems(parent *fyne.MenuItem) {
	recentlyOpenedFiles := fyne.CurrentApp().Preferences().StringList(RecentlyOpened)

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
	decodeHistoryData := []string{}

	var l *widget.List
	l = widget.NewList(
		func() int { return len(decodeHistoryData) },
		func() fyne.CanvasObject {
			return NewHistoryItem("History item 0")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*HistoryItem).SetText(decodeHistoryData[lii])

			l.SetItemHeight(lii, co.(*HistoryItem).MinSize().Height)
		},
	)
	handleDecodeResult := func(result string) {
		decodeHistoryData = append(decodeHistoryData, result)
		l.Refresh()
	}

	l.Hide()
	// TODO: Fix the ugly double pointer situation with content
	// TODO: Finish placeholder implementation
	var content *container.Split
	mainWindow := &MainWindow{
		Window:          w,
		requestContent:  NewPacketDisplay(lang.L("Request"), w, handleDecodeResult),
		responseContent: NewPacketDisplay(lang.L("Response"), w, handleDecodeResult),
		analysisToolbar: NewAnalysisToolbar(filter, w, l, &content),
		PacketFilter:    filter,
		packetChan:      packetChan,
	}
	mainWindow.registerShortcuts(restart)
	mainWindow.makeMenu(func() { settings.MakeSettingsUI(w, restart).Show() })
	content = container.NewHSplit(
		container.NewVSplit(
			NewPacketList(mainWindow.PacketFilter, mainWindow),
			container.NewHSplit(
				mainWindow.requestContent,
				mainWindow.responseContent,
			),
		),
		l,
	)
	content.SetOffset(.80)
	mainWindow.SetContent(
		container.NewBorder(
			container.NewVBox(
				mainWindow.analysisToolbar,
				mainWindow.PacketFilter,
				widget.NewSeparator(),
			),
			nil,
			nil,
			nil,
			content,
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
			if m.analysisToolbar.IsRecording {
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
// If there is, it displays a confirmation dialog for the user to choose whether to load the
// crash data.
func (m *MainWindow) CheckForCrashData() {
	configDir, err := util.GetConfigDir()
	if err != nil {
		slog.Error("Error getting config location", "error", err)
	}
	crashData := filepath.Join(configDir, "crash.json")
	if _, err := os.Stat(crashData); err == nil {
		dialog.ShowConfirm(lang.L("Crash data found"), lang.L("There was crash data found. Want to load it?"), func(b bool) {
			if b {
				m.PacketFilter.LoadPacketsFromFile(crashData)
			}
			if err := os.Remove(crashData); err != nil {
				util.ReportUIError(err, m)
			}
		}, m)
	}
}
