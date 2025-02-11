package ui

import (
	"log/slog"

	"com.github.redawl.mitmproxy/packet"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowAndRun (packetChan chan packet.HttpPacket) {
    a := app.New()
    w := a.NewWindow("MITMProxy")
    w.Resize(fyne.NewSize(1440, 810))
    packetList := make([]*packet.HttpPacket, 0)
    content := widget.NewLabel("TEWDWDWDW")

    table := widget.NewTableWithHeaders(func() (rows int, cols int) {
        return len(packetList), 5
    }, func() fyne.CanvasObject {
        label := widget.NewLabel("")
        return label
    }, func(tci widget.TableCellID, co fyne.CanvasObject) {
        label := co.(*widget.Label)
        if packetList[tci.Row] != nil {
            p := packetList[tci.Row]

            switch(tci.Col) {
                case 0: label.SetText(p.Path)
                case 1: label.SetText(p.ClientIp)
                default: label.SetText("")
            }
            label.Refresh()
        }
    })

    table.OnSelected = func(id widget.TableCellID) {
        slog.Info("I ran", "packet", packetList[id.Row])
        content.SetText(string(packetList[id.Row].RespContent))
        content.Refresh()
        table.RefreshItem(id)
    }
    go func() {
        for {
            packet := <- packetChan
            packetList = append(packetList, &packet)
            table.Refresh()
        }
    }()

    packetListContainer := container.NewBorder(widget.NewLabel("Hello world"), nil, nil, nil, table)
    packetListContainer.Show()

    masterLayout := container.NewGridWithColumns(2, packetListContainer, container.NewScroll(content))
    w.SetContent(masterLayout)

    slog.Info("Showing ui")
    w.ShowAndRun()
}

