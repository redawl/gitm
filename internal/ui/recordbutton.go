package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

const (
	IS_RECORDING     = "Listening..."
	IS_NOT_RECORDING = "Stopped."
)

func (b *RecordButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewHBox(b.Button, b.Label))
}

// NewRecordButton creates a new RecordButton
func NewRecordButton() *RecordButton {
	button := &RecordButton{
		Button: &widget.Button{
			Text: "Record",
		},
		Label: widget.NewLabel(IS_NOT_RECORDING),
	}

	button.Button.OnTapped = func() {
		button.IsRecording = !button.IsRecording
		if button.IsRecording {
			button.Label.SetText(IS_RECORDING)
		} else {
			button.Label.SetText(IS_NOT_RECORDING)
		}
	}

	button.ExtendBaseWidget(button)

	return button
}
