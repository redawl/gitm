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
	Button *widget.Button
	Label  *widget.Label
	// IsRecording specified whether to record packets
	IsRecording bool
}

// NewRecordButton creates a new RecordButton
func NewRecordButton(packetFilter *PacketFilter, w fyne.Window) *RecordButton {
	isRecording := lang.L("Listening...")
	isNotRecording := lang.L("Stopped.")

	button := &RecordButton{
		Button: &widget.Button{
			Text: lang.L("Record"),
		},
		Label: &widget.Label{Text: isNotRecording},
	}

	button.Button.OnTapped = func() {
		startRecording := func() {
			button.IsRecording = !button.IsRecording
			if button.IsRecording {
				button.Label.SetText(isRecording)
				button.Label.Importance = widget.DangerImportance
				button.Label.Refresh()
			} else {
				button.Label.SetText(isNotRecording)
				button.Label.Importance = widget.MediumImportance
				button.Label.Refresh()
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
	return widget.NewSimpleRenderer(container.NewHBox(b.Button, b.Label))
}
