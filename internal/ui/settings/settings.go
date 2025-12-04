package settings

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal"
	"github.com/redawl/gitm/internal/util"
)

var _ fyne.Layout = (*entryLayout)(nil)

type entryLayout struct{}

const EntrySize = 400

func (l *entryLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	util.Assert(len(objects) == 1)

	return fyne.NewSize(EntrySize, objects[0].MinSize().Height)
}

func (l *entryLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	util.Assert(len(objects) == 1)

	objects[0].Resize(size)
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

func (t *tableLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	util.Assert(len(objects) == 1)

	return fyne.NewSize(EntrySize, t.height)
}

func (t *tableLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	util.Assert(len(objects) == 1)
	table := objects[0].(*widget.Table)

	spacerSize := theme.Size(theme.SizeNameSeparatorThickness)
	table.SetColumnWidth(0, (EntrySize-spacerSize)*0.4)
	table.SetColumnWidth(1, (EntrySize-spacerSize)*0.5)
	table.Resize(size)
}

func ipPortValidator(s string) error {
	if len(s) == 0 {
		return nil
	}

	parts := strings.Split(s, ":")

	if len(parts) != 2 {
		return fmt.Errorf("must have one colon")
	}

	ipParts := strings.Split(parts[0], ".")

	if len(ipParts) != 4 {
		return fmt.Errorf("only IPv4 addresses are supported for now")
	}

	for _, part := range ipParts {
		if i, err := strconv.Atoi(part); err != nil {
			return fmt.Errorf("parsing ip: %w", err)
		} else {
			if i < 0 || i > 255 {
				return fmt.Errorf("quad must be between 0 and 255")
			}
		}
	}

	if i, err := strconv.Atoi(parts[1]); err != nil {
		return fmt.Errorf("parsing port: %w", err)
	} else {
		if i < 0 || i > 65536 {
			return fmt.Errorf("port must be between 0 and 65536")
		}
	}

	return nil
}

func dirValidator(s string) error {
	if s == "" {
		return nil
	}
	if _, err := os.Stat(s); err != nil {
		return err
	}

	return nil
}

// MakeSettingsUI creates a window for settings that the user can modify
func MakeSettingsUI(w fyne.Window, restart func()) dialog.Dialog {
	a := fyne.CurrentApp()
	prefs := a.Preferences()

	socks5Url := &widget.Entry{
		Text:      prefs.String(internal.SocksListenURI),
		Validator: ipPortValidator,
	}
	pacURL := &widget.Entry{
		Text:      prefs.String(internal.PACListenURI),
		Validator: ipPortValidator,
	}
	pacEnabled := &widget.Check{
		Checked: prefs.Bool(internal.EnablePACServer),
		OnChanged: func(b bool) {
			if !b {
				pacURL.Disable()
			} else {
				pacURL.Enable()
			}
		},
	}

	pacEnabled.OnChanged(pacEnabled.Checked)

	configDir := &widget.Entry{
		Text:      prefs.String(internal.ConfigDir),
		Validator: dirValidator,
	}

	configDir.ActionItem = widget.NewButton(lang.L("Choose"), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				util.ReportUIErrorWithMessage("Error opening folder", err, w)
				return
			}
			if uri != nil {
				configDir.SetText(uri.Path())
			}
		}, w)
	})

	themeEntry := &widget.Entry{
		Text:      prefs.String(internal.Theme),
		Validator: dirValidator,
	}

	themeEntry.ActionItem = widget.NewButton(lang.L("Choose"), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				util.ReportUIErrorWithMessage("Error opening theme", err, w)
				return
			}
			if reader == nil {
				return
			}

			themeEntry.SetText(reader.URI().Path())
		}, w)
	})

	debugEnabled := &widget.Check{
		Checked: prefs.Bool(internal.EnableDebugLogging),
	}

	customDecodings := prefs.StringList(internal.CustomDecodings)

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
			entry.Validator = func(s string) error {
				if strings.Contains(s, ":") {
					return fmt.Errorf("cannot contain a colon")
				}

				return nil
			}
			return entry
		},
		func(id widget.TableCellID, co fyne.CanvasObject) {
			entry := co.(*widget.Entry)
			if id.Col == 0 {
				entry.OnChanged = func(s string) {
					decodingLabels[id.Row] = s
				}
				entry.SetText(decodingLabels[id.Row])
			} else {
				entry.OnChanged = func(s string) {
					decodingCommands[id.Row] = s
				}
				entry.SetText(decodingCommands[id.Row])
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
				template.(*widget.Label).SetText(lang.L("Label"))
			} else {
				template.(*widget.Label).SetText(lang.L("Command"))
			}
		}
	}

	table.Refresh()

	form := make([]*widget.FormItem, 0)
	// TODO: Remove entryLayout? How does this look now?
	// Keeping it causes issues on some devices
	form = append(form, widget.NewFormItem(lang.L("Socks5 Proxy URL"), socks5Url))
	form = append(form, widget.NewFormItem(lang.L("Enable PAC server"), pacEnabled))
	form = append(form, widget.NewFormItem(lang.L("PAC URL"), pacURL))
	form = append(form, widget.NewFormItem(lang.L("GITM Config Directory"), configDir))
	form = append(form, widget.NewFormItem(lang.L("Theme"), themeEntry))
	form = append(form, widget.NewFormItem(lang.L("Custom Decodings"),
		container.NewBorder(
			nil,
			container.NewHBox(
				widget.NewButton(lang.L("Add Decoding"), func() {
					decodingLabels = append(decodingLabels, "")
					decodingCommands = append(decodingCommands, "")
					table.Refresh()
				}),
			), nil, nil,
			NewTableLayout(table),
		),
	))

	form = append(form, widget.NewFormItem(lang.L("Enable Debug Logging"), debugEnabled))

	s := dialog.NewForm(
		lang.L("Settings"),
		lang.L("Save"),
		lang.L("Cancel"),
		form,
		func(b bool) {
			if b {
				prefs.SetString(internal.SocksListenURI, socks5Url.Text)
				prefs.SetBool(internal.EnablePACServer, pacEnabled.Checked)
				prefs.SetString(internal.PACListenURI, pacURL.Text)
				prefs.SetString(internal.ConfigDir, configDir.Text)
				if themeEntry.Text == "" {
					fyne.CurrentApp().Settings().SetTheme(nil)
				} else if reader, err := os.Open(themeEntry.Text); err != nil {
					util.ReportUIErrorWithMessage(lang.L("Error open theme"), err, w)
				} else if th, err := theme.FromJSONReader(reader); err != nil {
					util.ReportUIErrorWithMessage(lang.L("Error parsing theme"), err, w)
				} else {
					fyne.CurrentApp().Settings().SetTheme(th)
				}
				prefs.SetString(internal.Theme, themeEntry.Text)
				prefs.SetBool(internal.EnableDebugLogging, debugEnabled.Checked)

				newCustomDecodings := make([]string, len(decodingLabels))

				for index := range decodingLabels {
					newCustomDecodings[index] = decodingLabels[index] + ":" + decodingCommands[index]
				}

				prefs.SetStringList(internal.CustomDecodings, newCustomDecodings)

				dialog.ShowConfirm(lang.L("Success!"), lang.L("New settings saved, would you like to restart the servers?"), func(b bool) {
					if b {
						restart()
					}
				}, w)
			} else {
				socks5Url.SetText(prefs.String(internal.SocksListenURI))
				debugEnabled.SetChecked(prefs.Bool(internal.EnableDebugLogging))
			}
		},
		w,
	)

	return s
}
