package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

// RecordButton is a button that allows the user to choose whether they want to record
// packets that are MITMed by the proxy.
type RecordButton struct {
	widget.BaseWidget
	// IsRecording specified whether to record packets
	IsRecording bool

	button *widget.Button
	label  *widget.Label
}

// NewRecordButton creates a new RecordButton
func NewRecordButton(packetFilter *PacketFilter, w fyne.Window) *RecordButton {
	isRecording := lang.L("Listening...")
	isNotRecording := lang.L("Stopped.")

	button := &RecordButton{
		button: &widget.Button{
			Text: lang.L("Record"),
		},
		label: &widget.Label{Text: isNotRecording},
	}

	button.button.OnTapped = func() {
		startRecording := func() {
			button.IsRecording = !button.IsRecording
			if button.IsRecording {
				button.label.SetText(isRecording)
				button.label.Importance = widget.DangerImportance
				button.label.Refresh()
			} else {
				button.label.SetText(isNotRecording)
				button.label.Importance = widget.MediumImportance
				button.label.Refresh()
			}
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

	button.ExtendBaseWidget(button)

	return button
}

func (b *RecordButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewHBox(b.button, b.label))
}
