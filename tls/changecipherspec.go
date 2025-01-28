package tls

import "log/slog"

type ChangeCipherSpecMessage struct {
    CSSProtocolType byte
}

func (message *ChangeCipherSpecMessage) GetLogAttrs () []slog.Attr {
    return []slog.Attr{
        slog.Any("CSSProtocolType", message.CSSProtocolType),
    }
}

func parseChangeCipherSpecMessage(protocolMessages []byte) ([]ProtocolMessage) {
    return []ProtocolMessage{
        &ChangeCipherSpecMessage{
            CSSProtocolType: protocolMessages[0],
        },
    }
}

