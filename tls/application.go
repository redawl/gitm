package tls

import "log/slog"

type ApplicationMessage struct {
    ApplicationData []byte
}

func (message *ApplicationMessage) GetLogAttrs () []slog.Attr {
    return []slog.Attr{
        slog.Any("ApplicationData", message.ApplicationData),
    }
}

func parseApplicationMessage (protocolMessage []byte) ([]ProtocolMessage) {
    return []ProtocolMessage{
        &ApplicationMessage{
            ApplicationData: protocolMessage,
        },
    }
}

