package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// RecordButton is a button that allows the user to choose whether they want to record
// packets that are MITMed by the proxy.
type RecordButton struct {
	widget.BaseWidget
	// IsRecording specified whether to record packets
	IsRecording  bool
	record, stop *widget.Button
}

// NewRecordButton creates a new RecordButton
func NewRecordButton(packetFilter *PacketFilter, w fyne.Window) *RecordButton {
	button := &RecordButton{
		record: &widget.Button{
			Text: lang.L("Record"),
			Icon: theme.Icon(theme.IconNameMediaPlay),
		},
		stop: &widget.Button{
			Text:       lang.L("Stop"),
			Importance: widget.DangerImportance,
			Icon:       theme.Icon(theme.IconNameMediaStop),
		},
	}

	button.record.Enable()
	button.stop.Disable()

	button.record.OnTapped = func() {
		startRecording := func() {
			button.IsRecording = true
			button.record.Disable()
			button.stop.Enable()
		}
		if len(packetFilter.Packets) > 0 {
			dialog.NewConfirm(
				lang.L("Overwrite packets"),
				lang.L("Starting a new capture will overwrite existing packets. Are you sure?"),
				func(b bool) {
					if b {
						packetFilter.ClearPackets()
						startRecording()
					}
				},
				w).Show()
		} else {
			startRecording()
		}
	}

	button.stop.OnTapped = func() {
		button.IsRecording = false
		button.record.Enable()
		button.stop.Disable()
	}

	button.ExtendBaseWidget(button)

	return button
}

func (b *RecordButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewHBox(b.record, b.stop))
}
