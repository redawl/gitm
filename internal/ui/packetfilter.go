package ui

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/util"
)

// PacketFilter is a text input that allows the user to filter the
// packets captured by the proxy.
type PacketFilter struct {
	widget.BaseWidget
	entry           *widget.Entry
	parent          fyne.Window
	Packets         []packet.Packet
	filteredPackets []packet.Packet
	listeners       []func()
}

// NewPacketFilter creates a new PacketFilter
func NewPacketFilter(w fyne.Window) *PacketFilter {
	prefs := fyne.CurrentApp().Preferences()
	input := &PacketFilter{
		entry: &widget.Entry{
			Text: prefs.String("PacketFilter"),
		},
		Packets: make([]packet.Packet, 0),
		parent:  w,
	}

	input.entry.OnChanged = func(s string) {
		prefs.SetString("PacketFilter", s)
		input.triggerListeners()
	}

	input.AddListener(func() {
		input.filteredPackets = filterPackets(input.entry.Text, input.Packets)
	})

	input.ExtendBaseWidget(input)

	return input
}

func (p *PacketFilter) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewBorder(
			nil,
			nil,
			widget.NewLabel(lang.L("Filter packets")),
			nil,
			p.entry,
		),
	)
}

// AppendPacket appends packet to the list trackets by p
// Calls all listeners added by AddListener
func (p *PacketFilter) AppendPacket(packet packet.Packet) {
	p.Packets = append(p.Packets, packet)
	p.triggerListeners()
}

// SetPackets overwrites the tracked packets with packets
// Calls all listeners added by AddListener
func (p *PacketFilter) SetPackets(newPackets []packet.Packet) {
	p.Packets = newPackets
	slices.SortFunc(p.Packets, func(a, b packet.Packet) int {
		return a.TimeStamp().Compare(b.TimeStamp())
	})
	p.triggerListeners()
}

// FindPacket searches the tracked packets for a matching packet
func (p *PacketFilter) FindPacket(httpPacket packet.Packet) packet.Packet {
	return httpPacket.FindPacket(p.Packets)
}

// ClearPackets resets the list of tracked packets
// Calls all listeners added by AddListener
func (p *PacketFilter) ClearPackets() {
	p.Packets = make([]packet.Packet, 0)
	p.triggerListeners()
}

// SavePackets asks the user for a file to save to,
// and then json marshalls the packet list, saving the result to the file.
func (p *PacketFilter) SavePackets() {
	jsonString, err := packet.MarshalPackets(p.Packets)
	if err != nil {
		util.ReportUiErrorWithMessage("Error marshalling packetList", err, p.parent)
		return
	}

	dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			util.ReportUiErrorWithMessage("Error saving to file", err, p.parent)
			return
		}

		if writer == nil {
			return
		}
		defer writer.Close() // nolint:errcheck

		if _, err := writer.Write(jsonString); err != nil {
			util.ReportUiErrorWithMessage("Error saving to file", err, p.parent)
			return
		}

		dialog.NewInformation(lang.L("Success!"), fmt.Sprintf(lang.L("Saved packets to %s successfully."), writer.URI().Path()), p.parent).Show()
	}, p.parent).Show()
}

// LoadPackets asks the user for a file to load from
// and then loads packets from that file
func (p *PacketFilter) LoadPackets() {
	showFilePicker := func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				util.ReportUiErrorWithMessage("Error reading from file", err, p.parent)
				return
			}

			if reader == nil {
				return
			}

			currRecentlyOpened := fyne.CurrentApp().Preferences().StringList(RECENTLY_OPENED)
			length := min(len(currRecentlyOpened)+1, 10)
			recentlyOpened := make([]string, length)

			recentlyOpened[0] = reader.URI().Path()
			count := 1
			for i := 1; i < length; i++ {
				if currRecentlyOpened[i-1] != recentlyOpened[0] {
					recentlyOpened[i] = currRecentlyOpened[i-1]
					count++
				}
			}

			fyne.CurrentApp().Preferences().SetStringList(RECENTLY_OPENED, recentlyOpened[:count])
			p.LoadPacketsFromReader(reader)
		}, p.parent).Show()
	}

	if len(p.Packets) > 0 {
		dialog.NewConfirm(
			lang.L("Overwrite packets"),
			lang.L("Are you sure you want to overwrite the currently displayed packets?"),
			func(confirmed bool) {
				if confirmed {
					showFilePicker()
				}
			},
			p.parent).Show()
	} else {
		showFilePicker()
	}
}

// LoadPacketsFromMostRecentFile loads packets from the file most recently opened
func (p *PacketFilter) LoadPacketsFromFile(filename string) {
	reader, err := os.Open(filename)
	if err != nil {
		dialog.ShowError(err, p.parent)
		return
	}

	p.LoadPacketsFromReader(reader)
}

// LoadPacketsFromReader json unmarshals the reader contents
func (p *PacketFilter) LoadPacketsFromReader(reader io.Reader) {
	fileContents, err := io.ReadAll(reader)
	if err != nil {
		util.ReportUiError(err, p.parent)
		return
	}

	packets := make([]packet.Packet, 0)
	if err := packet.UnmarshalPackets(fileContents, &packets); err != nil {
		util.ReportUiError(err, p.parent)
		return
	}

	p.SetPackets(packets)
}

// FilteredPackets returns the list of packets that match the current filter
// input by the user
func (p *PacketFilter) FilteredPackets() []packet.Packet {
	return p.filteredPackets
}

// AddListener adds a listener function that will be called by p whenever the
// tracked packet list changes
func (p *PacketFilter) AddListener(l func()) {
	p.listeners = append(p.listeners, l)
}

func (p *PacketFilter) triggerListeners() {
	for _, l := range p.listeners {
		fyne.Do(l)
	}
}

func getTokens(filterString string) []internal.FilterToken {
	filterStringStripped := strings.Trim(filterString, " ")

	tokens := make([]internal.FilterToken, 0)

	i := 0
	length := len(filterStringStripped)

	for i < length {
		if filterStringStripped[i] == ' ' {
			i++
			continue
		}

		token := internal.FilterToken{}
		colonIndex := strings.Index(filterStringStripped[i:], ":")

		if colonIndex == -1 {
			return tokens
		} else {
			colonIndex += i
		}

		// Get filter type
		token.FilterType = filterStringStripped[i:colonIndex]
		if strings.Contains(token.FilterType, " ") {
			// Get rid of prev cruft
			spaceIndex := strings.Index(token.FilterType, " ")
			token.FilterType = token.FilterType[spaceIndex+1:]
		}

		if len(filterStringStripped) <= colonIndex+1 {
			// found filterType without filterContent
			token.Negate = false
			tokens = append(tokens, token)
			return tokens
		}

		// Get filter content
		if filterStringStripped[colonIndex+1] == '-' {
			token.Negate = true
			colonIndex++
		} else {
			token.Negate = false
		}

		if len(filterStringStripped) <= colonIndex+1 {
			// found filterType without filterContent
			tokens = append(tokens, token)
			return tokens
		}

		if filterStringStripped[colonIndex+1] == '"' {
			quoteIndex := strings.Index(filterStringStripped[colonIndex+2:], "\"")
			if quoteIndex == -1 {
				spaceIndex := strings.Index(filterStringStripped[colonIndex+2:], " ")

				if spaceIndex == -1 {
					token.FilterContent = filterStringStripped[colonIndex+1:]
					i = length
				} else {
					token.FilterContent = filterStringStripped[colonIndex+1 : colonIndex+spaceIndex]
					i = spaceIndex + colonIndex
				}
			} else {
				token.FilterContent = filterStringStripped[colonIndex+2 : colonIndex+quoteIndex+2]
				i = quoteIndex + colonIndex + 1
			}
		} else {
			spaceIndex := strings.Index(filterStringStripped[colonIndex:], " ")

			if spaceIndex == -1 {
				token.FilterContent = filterStringStripped[colonIndex+1:]
				i = length
			} else {
				token.FilterContent = filterStringStripped[colonIndex+1 : colonIndex+spaceIndex]
				i = spaceIndex + colonIndex
			}
		}

		tokens = append(tokens, token)
	}

	return tokens
}

func filterPackets(filterString string, packets []packet.Packet) []packet.Packet {
	filterPairs := getTokens(filterString)
	passedPackets := make([]packet.Packet, 0, len(packets))

	for _, p := range packets {
		if p.MatchesFilter(filterPairs) {
			passedPackets = append(passedPackets, p)
		}
	}

	return passedPackets
}
