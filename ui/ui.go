package ui

import (
	"log/slog"

	"com.github.redawl.mitmproxy/packet"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowAndRun (packetChan chan packet.HttpPacket) {
    a := app.New()
    w := a.NewWindow("MITMProxy")
    w.Resize(fyne.NewSize(1440, 810))
    packetList := make([]*packet.HttpPacket, 0)
    content := widget.NewMultiLineEntry()
    content.TextStyle = fyne.TextStyle{
        Monospace: true,
    }

    table := widget.NewList(func() (int) {
        return len(packetList)
    }, func() fyne.CanvasObject {
        label := canvas.NewText("", nil)
        return label
    }, func(li widget.ListItemID, co fyne.CanvasObject) {
        label := co.(*canvas.Text)
        if packetList[li] != nil {
            p := packetList[li]
            encoding := p.RespHeaders["Content-Encoding"]
            if len(encoding) > 0 {
                label.Text = p.Path + " - " + p.RespHeaders["Content-Encoding"][0]
            } else {
                label.Text = p.Path
            }
            label.Refresh()
        }
    })

    table.OnSelected = func(id widget.ListItemID) {
        content.SetText(string(packetList[id].RespContent))
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

    masterLayout := container.NewAdaptiveGrid(2, packetListContainer, container.NewScroll(content))
    w.SetContent(masterLayout)

    slog.Info("Showing ui")
    w.ShowAndRun()
}

