package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
	"github.com/redawl/gitm/internal/ui/settings"
)

func makeMenu(clearHandler func(), saveHandler func(), loadHandler func(), settingsHandler func()) *fyne.MainMenu {
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Load", loadHandler),
			fyne.NewMenuItem("Clear", clearHandler),
			fyne.NewMenuItem("Save", saveHandler),
			fyne.NewMenuItem("Settings", settingsHandler),
		),
		MakeHelp(),
	)
	return mainMenu
}

// MakeUi Creates the Fyne UI for GITM, and then runs the UI event loop.
func MakeUi(packetChan chan packet.HttpPacket, restart func()) fyne.Window {
	a := fyne.CurrentApp()

	recordButton := NewRecordButton()

	w := a.NewWindow("Gopher in the middle")
	w.SetMaster()

	// TODO: Remove hardcoded default size.
	// We'll need to figure out how we want the application to look when
	// first opened. It would be nice to simulate a "Windowed fullscreen/borderless" look,
	// but fyne does not have direct support.
	w.Resize(fyne.NewSize(1920, 1080))

	requestContent := NewPacketDisplay("Request", w)
	responseContent := NewPacketDisplay("Response", w)

	packetFilter := NewPacketFilter()

	uiList := widget.NewList(func() int {
		return len(packetFilter.FilteredPackets())
	}, func() fyne.CanvasObject {
		return NewPacketRow()
	}, func(li widget.ListItemID, co fyne.CanvasObject) {
		row := co.(*PacketRow)
		filteredPackets := packetFilter.FilteredPackets()
		if li < len(filteredPackets) && filteredPackets[li] != nil {
			p := filteredPackets[li]
			row.UpdateRow(*p)
		}
	})
	uiList.HideSeparators = true

	packetFilter.AddListener(uiList.Refresh)

	uiList.OnSelected = func(id widget.ListItemID) {
		filteredPackets := packetFilter.FilteredPackets()
		requestContent.SetText(FormatRequestContent(filteredPackets[id]))
		responseContent.SetText(FormatResponseContent(filteredPackets[id]))
	}

	go func() {
		for {
			p := <-packetChan
			if recordButton.IsRecording {
				existingPacket := packetFilter.FindPacket(&p)

				if existingPacket != nil {
					existingPacket.UpdatePacket(&p)
				} else {
					packetFilter.AppendPacket(&p)
				}

				fyne.Do(uiList.Refresh)
			}
		}
	}()

	packetListContainer := container.NewBorder(
		container.NewVBox(
			container.NewHBox(recordButton),
			container.NewBorder(
				nil,
				nil,
				widget.NewLabel("Filter packets"),
				nil,
				packetFilter,
			),
		),
		nil,
		nil,
		nil,
		uiList,
	)

	masterLayout := container.NewVSplit(packetListContainer,
		container.NewHSplit(
			requestContent,
			responseContent,
		),
	)

	w.SetMainMenu(
		makeMenu(
			packetFilter.ClearPackets,
			func() {
				jsonString, err := json.Marshal(packetFilter.Packets)
				if err != nil {
					slog.Error("Error marshalling packetList", "error", err)
					dialog.ShowError(err, w)
					return
				}

				dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						slog.Error("Error saving to file", "filename", writer.URI().Path(), "error", err)
						dialog.ShowError(err, w)
						return
					}

					if writer == nil {
						return
					}
					defer writer.Close() // nolint:errcheck

					if _, err := writer.Write(jsonString); err != nil {
						slog.Error("Error saving to file", "filename", writer.URI().Path(), "error", err)
						dialog.ShowError(err, w)
						return
					}

					dialog.NewInformation("Success!", fmt.Sprintf("Saved packets to %s successfully.", writer.URI().Path()), w).Show()
				}, w).Show()
			},
			func() {
				showConfirmDialog := func() {
					dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
						if err != nil {
							slog.Error("Error saving to file", "filename", reader.URI().Path(), "error", err)
							dialog.ShowError(err, w)
							return
						}

						if reader == nil {
							return
						}

						fileContents, err := io.ReadAll(reader)
						if err != nil {
							slog.Error("Error reading from file", "filename", reader.URI().Path(), "error", err)
							dialog.ShowError(err, w)
							return
						}

						packets := make([]*packet.HttpPacket, 0)
						if err := json.Unmarshal(fileContents, &packets); err != nil {
							slog.Error("Error unmarshalling file contents", "filename", reader.URI().Path(), "error", err)
							dialog.ShowError(err, w)
							return
						}

						packetFilter.SetPackets(packets)
					}, w).Show()
				}

				if len(packetFilter.Packets) > 0 {
					dialog.NewConfirm(
						"Overwrite packets",
						"Are you sure you want to overwrite the currently displayed packets?",
						func(confirmed bool) {
							if confirmed {
								showConfirmDialog()
							}
						},
						w).Show()
				} else {
					showConfirmDialog()
				}
			},
			func() {
				settings.MakeSettingsUi(restart).Show()
			},
		),
	)

	w.SetContent(masterLayout)

	return w
}
