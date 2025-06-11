package settings

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestMakeSettings(t *testing.T) {
	_ = test.NewApp()

	settingsWindow := MakeSettingsUi(func() {})

	test.AssertRendersToImage(t, "settings.png", settingsWindow.Canvas())
}
