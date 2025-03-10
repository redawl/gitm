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

func MakeSettingsUi(a fyne.App) {
    w := a.NewWindow("Settings")
    prefs := a.Preferences()
    header := &canvas.Text{
        Text: "GITM Settings",
        TextSize: 30,
    }

    socks5Url := &widget.Entry{
        Text: prefs.String(config.SOCKS_LISTEN_URI),
    }

    cacertUrl := &widget.Entry{
        Text: prefs.String(config.CACERT_LISTEN_URI),
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
        widget.NewLabel("Socks5 proxy URL"), socks5Url,
        widget.NewLabel("Cacert proxy URL"), cacertUrl,
        widget.NewLabel("HTTP proxy URL"), httpUrl,
        widget.NewLabel("HTTPS proxy URL"), httpsUrl,
        widget.NewLabel("Enable debug logging"), debugEnabled,
    )

    form.Resize(fyne.NewSize(form.Size().Width + 400, form.Size().Height))

    formcontrols := container.NewHBox(widget.NewButton("Save", func() {
        prefs.SetString(config.SOCKS_LISTEN_URI, socks5Url.Text)
        prefs.SetString(config.CACERT_LISTEN_URI, cacertUrl.Text)
        prefs.SetString(config.HTTP_LISTEN_URI, httpUrl.Text)
        prefs.SetString(config.TLS_LISTEN_URI, httpsUrl.Text)
        prefs.SetBool(config.ENABLE_DEBUG_LOGGING, debugEnabled.Checked)

        successPopup := dialog.NewInformation("Success!", "New settings saved, but won't be applied until you restart GITM", w)
        successPopup.Show()
    }), widget.NewButton("Reset",func() {
        socks5Url.SetText(prefs.String(config.SOCKS_LISTEN_URI))    
        cacertUrl.SetText(prefs.String(config.CACERT_LISTEN_URI))
        httpUrl.SetText(prefs.String(config.HTTP_LISTEN_URI))
        httpsUrl.SetText(prefs.String(config.TLS_LISTEN_URI))
        debugEnabled.SetChecked(prefs.Bool(config.ENABLE_DEBUG_LOGGING))
    }))

    w.SetContent(container.NewVBox(header, form, formcontrols))

    w.Show()
}

