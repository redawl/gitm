package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

var SaveShortcut fyne.Shortcut = &desktop.CustomShortcut{KeyName: "S", Modifier: fyne.KeyModifierControl}

// registerShortcuts registers the toplevel shortcuts for gitm
func registerShortcuts(p *PacketFilter, w fyne.Window) {
	c := w.Canvas()

	c.AddShortcut(SaveShortcut, func(shortcut fyne.Shortcut) { p.SavePackets() })
}
