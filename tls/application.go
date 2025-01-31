package tls

import (
	"encoding/json"
	"log/slog"
)

type ApplicationMessage struct {
    ApplicationData []byte
}

func (message ApplicationMessage) GetLogAttrs () []slog.Attr {
    // Commented out for now; no need to see encrypted data
    // return []slog.Attr{
    //     slog.Any("ApplicationData", message.ApplicationData),
    // }

    return []slog.Attr{}
}

func (message ApplicationMessage) MarshalJSON () ([]byte, error) {
    valueMap := make(map[string]any)
    valueMap["Applicationdata"] = message.ApplicationData
    return json.Marshal(valueMap)
}

func parseApplicationMessage (protocolMessage []byte) ([]ProtocolMessage) {
    return []ProtocolMessage{
        &ApplicationMessage{
            ApplicationData: protocolMessage,
        },
    }
}

