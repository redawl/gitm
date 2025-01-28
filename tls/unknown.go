package tls

import "log/slog"

type UnknownMessage struct {
    Tag string
    ProtocolMessages []byte
}

func (message *UnknownMessage) GetLogAttrs () []slog.Attr {
    return []slog.Attr{
        slog.String("Tag", message.Tag),
        slog.Any("ProtocolMessages", message.ProtocolMessages),
    }
}
