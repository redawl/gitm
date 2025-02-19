package ui

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
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
    } else if len(path) > 100 {
        path = path[:100] + "..."
    }

    row.HttpLine.Text = fmt.Sprintf("%s %s %s - %s %s", p.Method, path, p.ReqProto, p.RespProto, p.Status)
    
    row.HttpLine.Refresh()

    request := fmt.Sprintf(
        "%s %s %s\n%s\n%s", 
        p.Method, 
        "/" + strings.Join(strings.Split(p.Path, "/")[1:], "/"),
        p.ReqProto, 
        formatHeaders(p.ReqHeaders), decodeBody(p.ReqContent, p.ReqHeaders["Content-Encoding"], p.ReqHeaders["Content-Type"]),
    )

    response := fmt.Sprintf(
        "%s %s\n%s\n%s", 
        p.RespProto, 
        p.Status, 
        formatHeaders(p.RespHeaders), decodeBody(p.RespContent, p.RespHeaders["Content-Encoding"], p.ReqHeaders["Content-Type"]),
    )

    row.ViewReq.OnTapped = func() {
        content.SetText(request)
        content.Refresh()
    }

    row.ViewResp.OnTapped = func() {
        content.SetText(response)
        content.Refresh()
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

func decodeBody(body []byte, contentEncodings []string, contentTypes []string) string {
    ret := body
    if len(contentEncodings) > 0 {
        decoded := bytes.NewReader(body)
        for _, contentEncoding := range contentEncodings {
            switch contentEncoding {
                case "gzip": {
                    decoded, err := gzip.NewReader(decoded)

                    if err != nil {
                        slog.Error("Failed decoding gzip", "error", err)
                        break
                    }

                    ret, err = io.ReadAll(decoded)

                    if err != nil {
                        slog.Error("Failed reading stream", "error", err)
                        break
                    }
                }
                case "deflate": {
                    decoded := flate.NewReader(decoded)

                    var err error
                    ret, err = io.ReadAll(decoded)

                    if err != nil {
                        slog.Error("Failed reading stream", "error", err) 
                        break
                    }
                }
                case "UTF-8":
                case "none":
                default: {
                    slog.Error("Unhandled compression", "compression", contentEncoding)
                    break
                }
            }
        }
    }

    if len(contentTypes) > 0 {
        for _, contentType := range contentTypes {
            switch contentType {
                case "application/json": {
                    buff := bytes.NewBuffer([]byte{})
                    err := json.Compact(buff, ret)

                    if err != nil {
                        slog.Error("Error compacting json", "error", err, "content", ret)
                        break
                    }

                    err = json.Indent(buff, buff.Bytes(), "\t", "")

                    if err != nil {
                        slog.Error("Error indenting json", "error", err, "content", ret)
                        break
                    }

                    ret, err = io.ReadAll(buff)

                    if err != nil {
                        slog.Error("Error reading indented json", "error", err)
                        break
                    }
                }
                case "text/html":
                default: {
                    slog.Error("Unhandled content type", "content-type", contentType)
                }
            }
        }
    }

    return string(ret)
}
