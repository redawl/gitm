package settings

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/config"
	"github.com/redawl/gitm/internal/util"
)

var _ fyne.Layout = (*entryLayout)(nil)

type entryLayout struct{}

const ENTRY_SIZE = 400

func (l *entryLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	util.Assert(len(objs) == 1)

	return fyne.NewSize(ENTRY_SIZE, objs[0].MinSize().Height)
}

func (l *entryLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	util.Assert(len(objs) == 1)

	objs[0].Resize(size)
}

var _ fyne.Layout = (*tableLayout)(nil)

type tableLayout struct {
	height float32
}

func NewTableLayout(table *widget.Table) *fyne.Container {
	tl := &tableLayout{}
	separatorSize := theme.Size(theme.SizeNameSeparatorThickness)
	entryHeight := widget.NewEntry().MinSize().Height
	labelHeight := widget.NewLabel("").MinSize().Height
	padding := theme.Size(theme.SizeNamePadding)

	tl.height = (4 * (padding - separatorSize)) + (labelHeight) + (4 * entryHeight)
	return container.New(tl, table)
}

func (t *tableLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	util.Assert(len(objs) == 1)

	return fyne.NewSize(ENTRY_SIZE, t.height)
}

func (t *tableLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	util.Assert(len(objs) == 1)
	table := objs[0].(*widget.Table)

	spacerSize := theme.Size(theme.SizeNameSeparatorThickness)
	table.SetColumnWidth(0, (ENTRY_SIZE-spacerSize)*0.4)
	table.SetColumnWidth(1, (ENTRY_SIZE-spacerSize)*0.5)
	table.Resize(size)
}

// MakeSettingsUi creates a window for settings that the user can modify
func MakeSettingsUi(restart func()) fyne.Window {
	a := fyne.CurrentApp()

	for _, window := range a.Driver().AllWindows() {
		if window.Title() == "Settings" {
			window.RequestFocus()
			return window
		}
	}

	w := a.NewWindow("Settings")
	prefs := a.Preferences()
	header := container.NewPadded(&widget.Label{
		Text:     "GITM Settings",
		SizeName: theme.SizeNameHeadingText,
	})

	socks5Url := &widget.Entry{
		Text: prefs.String(config.SOCKS_LISTEN_URI),
	}

	httpUrl := &widget.Entry{
		Text: prefs.String(config.HTTP_LISTEN_URI),
	}

	httpsUrl := &widget.Entry{
		Text: prefs.String(config.TLS_LISTEN_URI),
	}

	debugEnabled := &widget.Check{
		Checked: prefs.Bool(config.ENABLE_DEBUG_LOGGING),
	}

	customDecodings := prefs.StringList(config.CUSTOM_DECODINGS)

	decodingLabels := make([]string, len(customDecodings))
	decodingCommands := make([]string, len(customDecodings))

	for index, decoding := range customDecodings {
		decodingIndex := strings.Index(decoding, ":")
		label, command := decoding[:decodingIndex], decoding[decodingIndex+1:]
		decodingLabels[index] = label
		decodingCommands[index] = command
	}

	table := widget.NewTable(
		func() (int, int) { return len(decodingLabels), 2 },
		func() fyne.CanvasObject {
			entry := widget.NewEntry()
			return entry
		},
		func(tci widget.TableCellID, co fyne.CanvasObject) {
			entry := co.(*widget.Entry)
			if tci.Col == 0 {
				entry.OnChanged = func(s string) {
					decodingLabels[tci.Row] = s
				}
				entry.SetText(decodingLabels[tci.Row])
			} else {
				entry.OnChanged = func(s string) {
					decodingCommands[tci.Row] = s
				}
				entry.SetText(decodingCommands[tci.Row])
			}
		},
	)

	table.HideSeparators = true
	table.ShowHeaderRow = true
	table.CreateHeader = func() fyne.CanvasObject {
		label := widget.NewLabel("")

		label.TextStyle.Bold = true
		label.Alignment = fyne.TextAlignCenter
		return label
	}

	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		if id.Row == -1 {
			if id.Col == 0 {
				template.(*widget.Label).SetText("Label")
			} else {
				template.(*widget.Label).SetText("Command")
			}
		}
	}

	table.Refresh()

	form := widget.NewForm()
	form.Append("Socks5 proxy URL", container.New(&entryLayout{}, socks5Url))
	form.Append("HTTP proxy URL", container.New(&entryLayout{}, httpUrl))
	form.Append("HTTPS proxy URL", container.New(&entryLayout{}, httpsUrl))
	form.Append("Custom decodings",
		container.NewBorder(
			nil,
			container.NewHBox(
				widget.NewButton("Add decoding", func() {
					decodingLabels = append(decodingLabels, "")
					decodingCommands = append(decodingCommands, "")
					table.Refresh()
				}),
			), nil, nil,
			NewTableLayout(table),
		),
	)

	form.Append("Enable debug logging", debugEnabled)

	form.SubmitText = "Save"
	form.OnSubmit = func() {
		prefs.SetString(config.SOCKS_LISTEN_URI, socks5Url.Text)
		prefs.SetString(config.HTTP_LISTEN_URI, httpUrl.Text)
		prefs.SetString(config.TLS_LISTEN_URI, httpsUrl.Text)
		prefs.SetBool(config.ENABLE_DEBUG_LOGGING, debugEnabled.Checked)

		newCustomDecodings := make([]string, len(decodingLabels))

		for index := range decodingLabels {
			newCustomDecodings[index] = decodingLabels[index] + ":" + decodingCommands[index]
		}

		prefs.SetStringList(config.CUSTOM_DECODINGS, newCustomDecodings)

		successPopup := dialog.NewConfirm("Success!", "New settings saved, would you like to restart the servers?", func(b bool) {
			if b {
				restart()
			}
		}, w)
		successPopup.Show()
	}

	form.CancelText = "Reset"
	form.OnCancel = func() {
		socks5Url.SetText(prefs.String(config.SOCKS_LISTEN_URI))
		httpUrl.SetText(prefs.String(config.HTTP_LISTEN_URI))
		httpsUrl.SetText(prefs.String(config.TLS_LISTEN_URI))
		debugEnabled.SetChecked(prefs.Bool(config.ENABLE_DEBUG_LOGGING))
	}
	form.Refresh()

	w.SetContent(container.NewVBox(header, form))

	return w
}
