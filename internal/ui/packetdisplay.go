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
        Entry: widget.Entry{
            MultiLine: true,
            Wrapping: fyne.TextWrapBreak,
            TextStyle: fyne.TextStyle{
                Monospace: true,
            },
        },
    }

    packetDisplay.ExtendBaseWidget(packetDisplay)

    packetDisplay.Show()

    return packetDisplay
}

func (p *PacketDisplay) TypedRune (r rune) {}

func (p *PacketDisplay) TypedKey (key *fyne.KeyEvent) {}
