package ui

import (
	"log/slog"
	"strings"

	"com.github.redawl.gitm/packet"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func makeMenu (recordHandler func(), clearHandler func()) *fyne.MainMenu {
    saveItem := fyne.NewMenuItem("Save", func() {
        
    })

    clearItem := fyne.NewMenuItem("Clear", clearHandler)

    recordItem := fyne.NewMenuItem("Record", recordHandler)
    fileMenu := fyne.NewMenu("File", clearItem, recordItem, saveItem)

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

    filterType := widget.NewSelect([]string{
        "Filter hostname",
        "Filter method",
        "Filter statuscode",
    }, func(s string){})

    filterType.SetSelectedIndex(0)

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
        packetList = filterPacketList(packetFullList, s, filterType.Selected)
        table.Refresh()
    }

    go func() {
        for {
            packet := <- packetChan
            if shouldRecord {
                packetFullList = append(packetFullList, &packet)
                packetList = filterPacketList(packetFullList, filterContent.Text, filterType.Selected)
                table.Refresh()
            }
        }
    }()

    packetListContainer := container.NewBorder(container.NewBorder(
        nil, nil, container.NewVBox(filterType, isRecording), nil, filterContent,
    ), nil, nil, nil, table)
    packetListContainer.Show()

    masterLayout := container.NewGridWithRows(2, packetListContainer, container.NewScroll(content))
    w.SetMainMenu(makeMenu(func() {
        shouldRecord = !shouldRecord
        if shouldRecord {
            isRecording.SetText("Recording: on")
        } else {
            isRecording.SetText("Recording: off")
        }

        isRecording.Refresh()
    }, func() {
        packetFullList = make([]*packet.HttpPacket, 0)
        packetList = make([]*packet.HttpPacket, 0)
    }))
    w.SetContent(masterLayout)

    slog.Info("Showing ui")
    w.ShowAndRun()
}

func filterPacketList(packetFullList []*packet.HttpPacket, filter string, filterType string) []*packet.HttpPacket {
    if len(filter) == 0 {
        return packetFullList 
    }

    packetList := []*packet.HttpPacket{}

    for _, p := range packetFullList {
        switch filterType {
            case "Filter hostname": {
                if strings.Contains(p.ServerIp, filter) {
                    packetList = append(packetList, p)
                }
            }
            case "Filter method": {
                if strings.Contains(p.Method, filter) {
                    packetList = append(packetList, p)
                }
            }
            case "Filter statuscode": {
                if strings.Contains(p.Status, filter) {
                    packetList = append(packetList, p)
                }
            }
        }
    }

    return packetList
}
