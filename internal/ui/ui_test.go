package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestMakeUi(t *testing.T) {
	_ = test.NewTempApp(t)

	window := MakeMainWindow(nil, nil)

	test.AssertRendersToImage(t, "mainWindow.png", window.Canvas())
}
