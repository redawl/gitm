package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
)

func TestPacketEntry_SelectedText(t *testing.T) {
	_ = test.NewApp()
	packetEntry := NewPacketEntry()
	packetEntry.SetText("Line1\nLine2\nLine3")

	// Simulate mouse moves
	packetEntry.MouseDown(&desktop.MouseEvent{
		Button: desktop.MouseButtonPrimary,
		PointEvent: fyne.PointEvent{
			Position: fyne.NewPos(0, 0),
		},
	})
	packetEntry.MouseMoved(&desktop.MouseEvent{
		Button: desktop.MouseButtonPrimary,
		PointEvent: fyne.PointEvent{
			Position: packetEntry.PositionForCursorLocation(0, 5),
		},
	})
	packetEntry.MouseUp(&desktop.MouseEvent{
		Button: desktop.MouseButtonPrimary,
		PointEvent: fyne.PointEvent{
			Position: packetEntry.PositionForCursorLocation(0, 5),
		},
	})

	expected := "Line1"
	actual := packetEntry.SelectedText()

	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestPacketEntry_SelectedTextScrolled(t *testing.T) {
	_ = test.NewApp()
	packetEntry := NewPacketEntry()
	packetEntry.Resize(fyne.NewSize(2, 2))
	packetEntry.SetText("Line1\nLine2\nLine3")
	packetEntry.ScrollToBottom()

	if packetEntry.HasSelectedText() {
		t.Errorf("Expected HasSelectedText() == false, got true")
	}

	// Simulate mouse moves
	packetEntry.MouseDown(&desktop.MouseEvent{
		Button: desktop.MouseButtonPrimary,
		PointEvent: fyne.PointEvent{
			Position: packetEntry.PositionForCursorLocation(2, 0),
		},
	})
	packetEntry.MouseMoved(&desktop.MouseEvent{
		Button: desktop.MouseButtonPrimary,
		PointEvent: fyne.PointEvent{
			Position: packetEntry.PositionForCursorLocation(2, 5),
		},
	})
	packetEntry.MouseUp(&desktop.MouseEvent{
		Button: desktop.MouseButtonPrimary,
		PointEvent: fyne.PointEvent{
			Position: packetEntry.PositionForCursorLocation(2, 5),
		},
	})

	expected := "Line3"
	actual := packetEntry.SelectedText()

	if expected != actual {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}
