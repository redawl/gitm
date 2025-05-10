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

func makeMenu (clearHandler func(), saveHandler func(), loadHandler func(), settingsHandler func(), getSelectedText func() string, w fyne.Window) *fyne.MainMenu {

    mainMenu := *fyne.NewMainMenu(
        fyne.NewMenu("File", 
            fyne.NewMenuItem("Load", loadHandler), 
            fyne.NewMenuItem("Clear", clearHandler), 
            fyne.NewMenuItem("Save", saveHandler), 
            fyne.NewMenuItem("Settings", settingsHandler),
        ), 
    )
    return &mainMenu
}

// ShowAndRun Creates the Fyne UI for GITM, and then runs the UI event loop.
func ShowAndRun (a fyne.App, packetChan chan packet.HttpPacket) {
    shouldRecord := false
    isRecording := widget.NewButton("Recording: off", func() {})

    isRecording.OnTapped = func() {
        shouldRecord = !shouldRecord
        if shouldRecord {
            isRecording.SetText("Recording: on")
        } else {
            isRecording.SetText("Recording: off")
        }

        isRecording.Refresh()
    }

    w := a.NewWindow("Gopher in the middle")
    w.SetMaster()
    w.Resize(fyne.NewSize(1920, 1080))

    packetFullList := make([]*packet.HttpPacket, 0)
    packetList := make([]*packet.HttpPacket, 0)
    requestContent := NewPacketDisplay("Request")
    responseContent := NewPacketDisplay("Response")

    filterContent := widget.NewEntry()
    filterContent.Text = a.Preferences().String("PacketFilter")

    uiList := widget.NewList(func() int {
        return len(packetList)
    }, func() fyne.CanvasObject {
        return NewPacketRow()
    }, func(li widget.ListItemID, co fyne.CanvasObject) {
        row := co.(*PacketRow)
        if li < len(packetList) && packetList[li] != nil {
            p := packetList[li]
            row.UpdateRow(*p)
        }
    })

    uiList.OnSelected = func(id widget.ListItemID) {
        requestContent.SetText(FormatRequestContent(packetList[id]))
        responseContent.SetText(FormatResponseContent(packetList[id]))
    }

    filterContent.OnChanged = func(s string) {
        a.Preferences().SetString("PacketFilter", s)
        packetList = FilterPackets(s, packetFullList)
        uiList.Refresh()
    }

    go func() {
        for {
            p := <- packetChan
            if shouldRecord {
                existingPacket := packet.FindPacket(&p, packetFullList)

                if existingPacket != nil {
                    existingPacket.UpdatePacket(&p)
                } else {
                    packetFullList = append(packetFullList, &p)
                    packetList = FilterPackets(filterContent.Text, packetFullList)
                }
                fyne.Do(uiList.Refresh)
            }
        }
    }()

    packetListContainer := container.NewBorder(container.NewVBox(
        container.NewHBox(isRecording), container.NewBorder(nil, nil, widget.NewLabel("Filter packets"), nil, filterContent),
    ), nil, nil, nil, uiList)

    masterLayout := container.NewVSplit(packetListContainer, 
        container.NewHSplit(
            container.NewScroll(requestContent),
            container.NewScroll(responseContent),
        ),
    )
    w.SetMainMenu(
        makeMenu(
            func() {
                packetFullList = make([]*packet.HttpPacket, 0)
                packetList = make([]*packet.HttpPacket, 0)
                uiList.Refresh()
            },
            func() {
                jsonString, err := json.Marshal(packetFullList)

                if err != nil {
                    slog.Error("Error marshalling packetList", "error", err)
                    dialog.ShowError(err, w)
                    return
                }

                saveFileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
                    if err != nil {
                        slog.Error("Error saving to file", "filename", writer.URI().Path(), "error", err)
                        dialog.ShowError(err, w)
                        return
                    }

                    if writer == nil {
                        return
                    }
                    _, err = writer.Write(jsonString)

                    if err != nil {
                        slog.Error("Error saving to file", "filename", writer.URI().Path(), "error", err)
                        dialog.ShowError(err, w)
                        return
                    }

                    successDialog := dialog.NewInformation("Success!", fmt.Sprintf("Saved packets to %s successfully.", writer.URI().Path()), w)
                    successDialog.Show()
                }, w)

                saveFileDialog.Show()
            },
            func() {
                showConfirmDialog := func() {
                    openFileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
                        if err != nil {
                            slog.Error("Error saving to file", "filename", reader.URI().Path(), "error", err)
                            dialog.ShowError(err, w)
                            return
                        }

                        if reader == nil {
                            return
                        }

                        packetFullList = make([]*packet.HttpPacket, 0)
                        fileContents, err := io.ReadAll(reader)

                        if err != nil {
                            slog.Error("Error reading from file", "filename", reader.URI().Path(), "error", err)
                            dialog.ShowError(err, w)
                            return
                        }
                        err = json.Unmarshal(fileContents, &packetFullList)

                        if err != nil {
                            slog.Error("Error unmarshalling file contents", "filename", reader.URI().Path(), "error", err)
                            dialog.ShowError(err, w)
                            return
                        }
                        
                        packetList = FilterPackets(filterContent.Text, packetFullList)
                        uiList.Refresh()
                    }, w)

                    openFileDialog.Show()
                }

                if len(packetFullList) > 0 {
                    confirmDialog := dialog.NewConfirm("Overwrite packets", "Are you sure you want to overwrite the currently displayed packets?", func(confirmed bool) {
                        if confirmed {
                            showConfirmDialog()
                        }
                    }, w)

                    confirmDialog.Show()
                } else {
                    showConfirmDialog()
                }
            },
            func() {
                settings.MakeSettingsUi(a)
            },
            func() string {
                if responseContent.HasSelectedText() {
                    return responseContent.SelectedText()
                } else if requestContent.HasSelectedText() {
                    return requestContent.SelectedText()
                }

                return ""
            },
            w,
        ),
    )

    w.SetContent(masterLayout)

    slog.Info("Showing ui")
    w.ShowAndRun()
}

