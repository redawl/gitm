package ui

import (
	"log/slog"

	"com.github.redawl.mitmproxy/packet"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func makeMenu () *fyne.MainMenu {

    file := fyne.NewMenu("File")

    mainMenu := *fyne.NewMainMenu(file)
    return &mainMenu
}

func ShowAndRun (packetChan chan packet.HttpPacket) {
    a := app.New()
    w := a.NewWindow("MITMProxy")
    packetList := make([]*packet.HttpPacket, 0)
    content := widget.NewMultiLineEntry()
    content.TextStyle = fyne.TextStyle{
        Monospace: true,
    }

    table := widget.NewList(func() int {
        return len(packetList)
    }, func() fyne.CanvasObject {
        return NewPacketRow()
    }, func(li widget.ListItemID, co fyne.CanvasObject) {
        row := co.(*PacketRow)
        if packetList[li] != nil {
            p := packetList[li]
            row.UpdateRow(*p, content)
        }
    })

    go func() {
        for {
            packet := <- packetChan
            packetList = append(packetList, &packet)
            table.Refresh()
        }
    }()

    packetListContainer := container.NewBorder(widget.NewLabel("MITMProxy"), nil, nil, nil, table)
    packetListContainer.Show()

    masterLayout := container.NewGridWithRows(2, packetListContainer, container.NewScroll(content))
    w.SetMainMenu(makeMenu())
    w.SetContent(masterLayout)

    slog.Info("Showing ui")
    w.ShowAndRun()
}

