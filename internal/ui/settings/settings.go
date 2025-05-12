package settings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/config"
)

type entryLayout struct {}
const ENTRY_SIZE = 400

func (ll *entryLayout) MinSize (objs []fyne.CanvasObject) fyne.Size {
    if len(objs) != 1 {
        panic("Entrylayout can only take 1 object!")
    }

    entry, ok := objs[0].(*widget.Entry)

    if !ok {
        panic("Entrylayout can only take entries")
    }

    return fyne.NewSize(ENTRY_SIZE, entry.MinSize().Height)
}

func (ll *entryLayout) Layout (objs []fyne.CanvasObject, size fyne.Size) {
    if len(objs) != 1 {
        panic("Entrylayout can only take 1 object!")
    }

    entry, ok := objs[0].(*widget.Entry)

    if !ok {
        panic("Entrylayout can only take entries")
    }

    entry.Resize(size)
}

func MakeSettingsUi(a fyne.App, restart func()) {
    w := a.NewWindow("Settings")
    prefs := a.Preferences()
    header := container.NewPadded(&canvas.Text{
        Text: "GITM Settings",
        TextSize: 30,
    })

    socks5Url := &widget.Entry{
        Text: prefs.String(config.SOCKS_LISTEN_URI),
    }

    httpUrl   := &widget.Entry{
        Text: prefs.String(config.HTTP_LISTEN_URI),
    }

    httpsUrl  := &widget.Entry{
        Text: prefs.String(config.TLS_LISTEN_URI),
    }

    debugEnabled := &widget.Check{
        Checked: prefs.Bool(config.ENABLE_DEBUG_LOGGING),
    }

    form := container.New(layout.NewFormLayout(),
        widget.NewLabel("Socks5 proxy URL"), container.New(&entryLayout{}, socks5Url),
        widget.NewLabel("HTTP proxy URL"), container.New(&entryLayout{}, httpUrl),
        widget.NewLabel("HTTPS proxy URL"), container.New(&entryLayout{}, httpsUrl),
        widget.NewLabel("Enable debug logging"), debugEnabled,
    )

    form.Resize(fyne.NewSize(form.Size().Width + ENTRY_SIZE, form.Size().Height))

    formcontrols := container.NewHBox(widget.NewButton("Save", func() {
        prefs.SetString(config.SOCKS_LISTEN_URI, socks5Url.Text)
        prefs.SetString(config.HTTP_LISTEN_URI, httpUrl.Text)
        prefs.SetString(config.TLS_LISTEN_URI, httpsUrl.Text)
        prefs.SetBool(config.ENABLE_DEBUG_LOGGING, debugEnabled.Checked)

        successPopup := dialog.NewConfirm("Success!", "New settings saved, would you like to restart the servers?", func(b bool) {
			if b {
				restart()
			}
		}, w)
        successPopup.Show()
    }), widget.NewButton("Reset",func() {
        socks5Url.SetText(prefs.String(config.SOCKS_LISTEN_URI))    
        httpUrl.SetText(prefs.String(config.HTTP_LISTEN_URI))
        httpsUrl.SetText(prefs.String(config.TLS_LISTEN_URI))
        debugEnabled.SetChecked(prefs.Bool(config.ENABLE_DEBUG_LOGGING))
    }))

    w.SetContent(container.NewVBox(header, form, formcontrols))

    w.Show()
}

