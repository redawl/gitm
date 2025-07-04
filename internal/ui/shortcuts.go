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
)

// registerShortcuts registers the toplevel shortcuts for gitm
func registerShortcuts(p *PacketFilter, w fyne.Window, restart func()) {
	c := w.Canvas()

	c.AddShortcut(SaveShortcut, func(shortcut fyne.Shortcut) { p.SavePackets() })
	c.AddShortcut(OpenShortcut, func(shortcut fyne.Shortcut) { p.LoadPackets() })
	c.AddShortcut(SettingsShortcut, func(shortcut fyne.Shortcut) { settings.MakeSettingsUi(restart) })
}
