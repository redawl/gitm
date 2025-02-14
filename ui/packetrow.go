package ui

import (
	"fmt"
	"log/slog"
	"strings"

	"com.github.redawl.mitmproxy/packet"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type PacketRow struct {
    widget.BaseWidget
    HttpLine canvas.Text
    ViewReq widget.Button
    ViewResp widget.Button
}

func defaultFunc() {
    slog.Debug("User clicked too fast!!!")
}

func NewPacketRow() *PacketRow {
    row := &PacketRow{
        HttpLine: *canvas.NewText("", nil),
        ViewReq: *widget.NewButton("View request", defaultFunc),
        ViewResp: *widget.NewButton("View response", defaultFunc),
    }

    row.HttpLine.TextStyle = fyne.TextStyle{
        Monospace: true,
        TabWidth: 4,
    }
    row.ExtendBaseWidget(row)

    return row
}

func (row *PacketRow) UpdateRow (p packet.HttpPacket, content *widget.Entry) {
    path := p.Path

    if len(path) == 0 {
        path = "/"
    }

    row.HttpLine.Text = fmt.Sprintf("%s %s HTTP/1.1", p.Method, path)
    row.HttpLine.Refresh()

    row.ViewReq.OnTapped = func() {
        content.SetText(
            row.HttpLine.Text + "\n" + formatHeaders(p.ReqHeaders) + "\n\n" + string(p.ReqContent),
        )
    }

    row.ViewResp.OnTapped = func() {
        content.SetText(
            fmt.Sprintf("HTTP/1.1 %d FIXME", p.Status) + "\n" + formatHeaders(p.RespHeaders) + "\n\n" + string(p.RespContent),
        )
    }
    
    row.ExtendBaseWidget(row)
    row.Refresh()
}

func (row *PacketRow) CreateRenderer () fyne.WidgetRenderer {
    c := container.NewBorder(nil, nil, nil, container.NewHBox(&row.ViewReq, &row.ViewResp), &row.HttpLine)
    return widget.NewSimpleRenderer(c)
}

func formatHeaders (headers map[string][]string) string {
    builder := strings.Builder{}

    for header, values := range headers {
        builder.WriteString(header + ": ") 
        for i, value := range values {
            if i == len(values) - 1 {
                builder.WriteString(value)
            } else {
                builder.WriteString(value + ", ")
            }
        }
        builder.WriteString("\n")
    }

    return builder.String()
}
