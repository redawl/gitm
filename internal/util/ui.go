package util

import (
	"fmt"

	"fyne.io/fyne/v2"
)

// NewWindowIfNotExists creates a window with the given title,
// unless there is already an existing window with the same title
// If a window with title already exists, that window is returned instead
// of a new window
func NewWindowIfNotExists(title string) fyne.Window {
	a := fyne.CurrentApp()

	for _, window := range a.Driver().AllWindows() {
		if window.Title() == title {
			window.RequestFocus()
			return window
		}
	}

	windowWidth := fmt.Sprintf("%s:WindowWidth", title)
	windowHeight := fmt.Sprintf("%s:WindowHeight", title)
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(
		float32(a.Preferences().FloatWithFallback(windowWidth, 1000)),
		float32(a.Preferences().FloatWithFallback(windowHeight, 800)),
	))

	w.SetOnClosed(func() {
		a.Preferences().SetFloat(windowWidth, float64(w.Canvas().Size().Width))
		a.Preferences().SetFloat(windowHeight, float64(w.Canvas().Size().Height))
	})

	return w
}
