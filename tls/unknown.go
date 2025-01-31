package tls

import (
	"encoding/json"
	"log/slog"
)

type UnknownMessage struct {
    Tag string
    ProtocolMessages []byte
}

func (message UnknownMessage) GetLogAttrs () []slog.Attr {
    return []slog.Attr{
        slog.String("Tag", message.Tag),
        slog.Any("ProtocolMessages", message.ProtocolMessages),
    }
}

func (message UnknownMessage) MarshalJSON() ([]byte, error) {
    return json.Marshal(message)
}
