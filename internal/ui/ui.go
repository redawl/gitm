package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
)

func makeMenu (clearHandler func(), saveHandler func(), loadHandler func()) *fyne.MainMenu {
    saveItem := fyne.NewMenuItem("Save", saveHandler)
    loadItem := fyne.NewMenuItem("Load", loadHandler)

    clearItem := fyne.NewMenuItem("Clear", clearHandler)

    fileMenu := fyne.NewMenu("File", loadItem, clearItem, saveItem)

    mainMenu := *fyne.NewMainMenu(fileMenu)
    return &mainMenu
}

func ShowAndRun (packetChan chan packet.HttpPacket) {
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
    a := app.New()
    w := a.NewWindow("Gopher in the middle")
    w.Resize(fyne.NewSize(1920, 1080))

    packetFullList := make([]*packet.HttpPacket, 0)
    packetList := make([]*packet.HttpPacket, 0)
    requestContent := NewPacketDisplay("Request")
    responseContent := NewPacketDisplay("Response")

    filterContent := widget.NewEntry()

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

                uiList.Refresh()
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
        ),
    )

    w.SetContent(masterLayout)

    slog.Info("Showing ui")
    w.ShowAndRun()
}

