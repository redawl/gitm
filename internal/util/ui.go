package util

import "fyne.io/fyne/v2"

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

	return a.NewWindow(title)
}
