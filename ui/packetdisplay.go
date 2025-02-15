package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type PacketDisplay struct {
    widget.Entry
}

func NewPacketDisplay() *PacketDisplay {
    return &PacketDisplay{
        Entry: *widget.NewMultiLineEntry(),
    }
}

func (p *PacketDisplay) TypedRune (r rune) {}

func (p *PacketDisplay) TypedKey (key *fyne.KeyEvent) {}
