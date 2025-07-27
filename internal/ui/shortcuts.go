package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/redawl/gitm/internal/ui/settings"
)

var (
	SaveShortcut     fyne.Shortcut = &desktop.CustomShortcut{KeyName: "S", Modifier: fyne.KeyModifierControl}
	OpenShortcut     fyne.Shortcut = &desktop.CustomShortcut{KeyName: "O", Modifier: fyne.KeyModifierControl}
	SettingsShortcut fyne.Shortcut = &desktop.CustomShortcut{KeyName: "S", Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}
	ClearShortcut    fyne.Shortcut = &desktop.CustomShortcut{KeyName: "X", Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}
	QuitShortcut     fyne.Shortcut = &desktop.CustomShortcut{KeyName: "Q", Modifier: fyne.KeyModifierControl}
)

// registerShortcuts registers the toplevel shortcuts for gitm
func (m *MainWindow) registerShortcuts(restart func()) {
	c := m.Canvas()

	c.AddShortcut(SaveShortcut, func(shortcut fyne.Shortcut) { m.PacketFilter.SavePackets() })
	c.AddShortcut(OpenShortcut, func(shortcut fyne.Shortcut) { m.PacketFilter.LoadPackets() })
	c.AddShortcut(SettingsShortcut, func(shortcut fyne.Shortcut) { settings.MakeSettingsUi(restart) })
	c.AddShortcut(ClearShortcut, func(shortcut fyne.Shortcut) { m.PacketFilter.ClearPackets() })
	c.AddShortcut(QuitShortcut, func(shortcut fyne.Shortcut) { fyne.CurrentApp().Quit() })
}
