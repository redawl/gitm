package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"com.github.redawl.gitm/packet"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func makeMenu (recordHandler func(), clearHandler func(), saveHandler func(), loadHandler func()) *fyne.MainMenu {
    saveItem := fyne.NewMenuItem("Save", saveHandler)
    loadItem := fyne.NewMenuItem("Load", loadHandler)

    clearItem := fyne.NewMenuItem("Clear", clearHandler)

    recordItem := fyne.NewMenuItem("Record", recordHandler)
    fileMenu := fyne.NewMenu("File", loadItem, clearItem, recordItem, saveItem)

    mainMenu := *fyne.NewMainMenu(fileMenu)
    return &mainMenu
}

func ShowAndRun (packetChan chan packet.HttpPacket) {
    shouldRecord := false
    isRecording := widget.NewLabel("Recording: off")
    a := app.New()
    w := a.NewWindow("GITM")

    packetFullList := make([]*packet.HttpPacket, 0)
    packetList := make([]*packet.HttpPacket, 0)
    content := widget.NewMultiLineEntry()
    content.Wrapping = fyne.TextWrapBreak
    content.TextStyle = fyne.TextStyle{
        Monospace: true,
    }

    filterContent := widget.NewEntry()

    table := widget.NewList(func() int {
        return len(packetList)
    }, func() fyne.CanvasObject {
        return NewPacketRow()
    }, func(li widget.ListItemID, co fyne.CanvasObject) {
        row := co.(*PacketRow)
        if li < len(packetList) && packetList[li] != nil {
            p := packetList[li]
            row.UpdateRow(*p, content)
        }
    })

    filterContent.OnChanged = func(s string) {
        packetList = FilterPackets(s, packetFullList)
        table.Refresh()
    }

    go func() {
        for {
            packet := <- packetChan
            if shouldRecord {
                packetFullList = append(packetFullList, &packet)
                packetList = FilterPackets(filterContent.Text, packetFullList)
                table.Refresh()
            }
        }
    }()

    packetListContainer := container.NewBorder(container.NewBorder(
        nil, nil, isRecording, nil, filterContent,
    ), nil, nil, nil, table)
    packetListContainer.Show()


    masterLayout := container.NewGridWithRows(2, packetListContainer, container.NewScroll(content))
    w.SetMainMenu(
        makeMenu(
            func() {
                shouldRecord = !shouldRecord
                if shouldRecord {
                    isRecording.SetText("Recording: on")
                } else {
                    isRecording.SetText("Recording: off")
                }

                isRecording.Refresh()
            }, 
            func() {
                packetFullList = make([]*packet.HttpPacket, 0)
                packetList = make([]*packet.HttpPacket, 0)
            },
            func() {
                jsonString, err := json.Marshal(packetList)

                if err != nil {
                    slog.Error("Error marshalling packetList", "error", err)
                    dialog.ShowError(err, w)
                    return
                }

                saveFileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
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
                openFileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
                    packetList = make([]*packet.HttpPacket, 0)
                    fileContents, err := io.ReadAll(reader)

                    if err != nil {
                        slog.Error("Error reading from file", "filename", reader.URI().Path(), "error", err)
                        dialog.ShowError(err, w)
                        return
                    }
                    err = json.Unmarshal(fileContents, &packetList)

                    if err != nil {
                        slog.Error("Error unmarshalling file contents", "filename", reader.URI().Path(), "error", err)
                        dialog.ShowError(err, w)
                        return
                    }
                    
                    packetFullList = packetList
                }, w)

                openFileDialog.Show()
            },
        ),
    )
    w.SetContent(masterLayout)

    slog.Info("Showing ui")
    w.ShowAndRun()
}

