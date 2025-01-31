package tls

import (
	"encoding/json"
	"log/slog"
)

type ChangeCipherSpecMessage struct {
    CSSProtocolType byte
}

func (message ChangeCipherSpecMessage) GetLogAttrs () []slog.Attr {
    return []slog.Attr{
        slog.Any("CSSProtocolType", message.CSSProtocolType),
    }
}

func (message ChangeCipherSpecMessage) MarshalJSON () ([]byte, error) {
    valueMap := make(map[string]any)
    valueMap["CSSProtocolType"] = message.CSSProtocolType
    return json.Marshal(valueMap)
}

func parseChangeCipherSpecMessage(protocolMessages []byte) ([]ProtocolMessage) {
    return []ProtocolMessage{
        &ChangeCipherSpecMessage{
            CSSProtocolType: protocolMessages[0],
        },
    }
}

