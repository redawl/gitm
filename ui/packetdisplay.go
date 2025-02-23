package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type PacketDisplay struct {
    widget.Entry
}

func NewPacketDisplay() *PacketDisplay {
    packetDisplay := &PacketDisplay{
        Entry: *widget.NewMultiLineEntry(),
    }
    packetDisplay.Wrapping = fyne.TextWrapBreak
    packetDisplay.TextStyle = fyne.TextStyle{
        Monospace: true,
    }

    return packetDisplay
}

func (p *PacketDisplay) TypedRune (r rune) {}

func (p *PacketDisplay) TypedKey (key *fyne.KeyEvent) {}
