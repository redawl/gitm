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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/redawl/gitm/internal/packet"
)

type PacketDisplay struct {
	widget.BaseWidget
	entry *PacketEntry
	label *widget.Label
}

func NewPacketDisplay(label string) *PacketDisplay {
	packetDisplay := &PacketDisplay{
		entry: NewPacketEntry(),
		label: &widget.Label{
			Text:     label,
			SizeName: theme.SizeNameSubHeadingText,
		},
	}

	packetDisplay.ExtendBaseWidget(packetDisplay)

	return packetDisplay
}

func (pd *PacketDisplay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewBorder(pd.label, nil, nil, nil, pd.entry))
}

func (pd *PacketDisplay) SetText(text string) {
	pd.entry.SetText(text)

	pd.entry.ScrollToTop()
}

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

func formatHeaders(headers map[string][]string) string {
	builder := strings.Builder{}

	for header, values := range headers {
		builder.WriteString(header + ": ")
		for i, value := range values {
			if i == len(values)-1 {
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
			case "gzip":
				{
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
			case "deflate":
				{
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
			default:
				{
					slog.Error("Unhandled compression", "compression", contentEncoding)
					break
				}
			}
		}
	}

	if json.Valid(ret) {
		buff := new(bytes.Buffer)
		err := json.Indent(buff, ret, "", "    ")

		if err != nil {
			slog.Error("Failed indenting json", "error", err)
		} else {
			ret = buff.Bytes()
		}
	}

	return string(ret)
}
