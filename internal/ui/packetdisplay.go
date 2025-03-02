package ui

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
)

type PacketDisplay struct {
    widget.BaseWidget
    Entry widget.Entry
    scrollContainer *container.Scroll
}

func NewPacketDisplay() *PacketDisplay {
    packetDisplay := &PacketDisplay{
        Entry: widget.Entry{
            MultiLine: true,
            Wrapping: fyne.TextWrapBreak,
            TextStyle: fyne.TextStyle{
                Monospace: true,
            },
        },
    }

    packetDisplay.scrollContainer = container.NewScroll(&packetDisplay.Entry)

    packetDisplay.ExtendBaseWidget(packetDisplay)

    return packetDisplay
}

func (pd *PacketDisplay) CreateRenderer() fyne.WidgetRenderer {
    return widget.NewSimpleRenderer(pd.scrollContainer)
}

func (pd *PacketDisplay) SetText(text string) {
    pd.Entry.SetText(text)
    pd.scrollContainer.ScrollToTop()
}

func (p *PacketDisplay) TypedRune (r rune) {}

func (p *PacketDisplay) TypedKey (key *fyne.KeyEvent) {}

func FormatRequestContent(p *packet.HttpPacket) string {
    return fmt.Sprintf(
        "%s %s %s\n%s\n%s", 
        p.Method, 
        p.Path,
        p.ReqProto, 
        formatHeaders(p.ReqHeaders),
        decodeBody(p.ReqBody, p.ReqHeaders["Content-Encoding"]),
    )
}

func FormatResponseContent(p *packet.HttpPacket) string {
    return fmt.Sprintf(
        "%s %s\n%s\n%s", 
        p.RespProto, 
        p.Status, 
        formatHeaders(p.RespHeaders),
        decodeBody(p.RespBody, p.RespHeaders["Content-Encoding"]),
    )
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

func decodeBody(body []byte, contentEncodings []string) string {
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

    return string(ret)
}

