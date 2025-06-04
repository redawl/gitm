package ui

import "fyne.io/fyne/v2/widget"

// RecordButton is a button that allows the user to choose whether they want to record
// packets that are MITMed by the proxy.
type RecordButton struct {
	widget.Button
	// IsRecording specified whether to record packets
	IsRecording bool
}

const (
	IS_RECORDING     = "Recording: on"
	IS_NOT_RECORDING = "Recording: off"
)

// NewRecordButton creates a new RecordButton
func NewRecordButton() *RecordButton {
	button := &RecordButton{
		Button: widget.Button{
			Text: IS_NOT_RECORDING,
		},
	}

	button.OnTapped = func() {
		button.IsRecording = !button.IsRecording
		if button.IsRecording {
			button.SetText(IS_RECORDING)
		} else {
			button.SetText(IS_NOT_RECORDING)
		}
	}

	button.ExtendBaseWidget(button)

	return button
}
