package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func MakeHelp() *fyne.Menu {
	app := fyne.CurrentApp()
	about := fyne.NewMenuItem("About", func() {
		w := app.NewWindow("About")

		w.SetContent(widget.NewLabel("Hello, about us!"))
		w.Show()
	})

	menu := fyne.NewMenu("Help", about)

	return menu
}
