package tls

import (
	"encoding/json"
	"log/slog"
)

// TODO: Implement level and description mappings
const (
    LevelWarning = 0x01
    LevelFatal   = 0x02
)

const (
    DescCloseNotify = 0
    DescUnexpectedMessage = 10
    DescBadRecordMac
)


type AlertMessage struct {
    Level byte
    Description byte
}

func (message AlertMessage) GetLogAttrs () []slog.Attr {
    return []slog.Attr{
        slog.Any("Level", message.Level),
        slog.Any("Description", message.Description),
    }
}

func (message AlertMessage) MarshalJSON() ([]byte, error) {
    valueMap := make(map[string]any)
    valueMap["Level"] = message.Level
    valueMap["Description"] = message.Description

    return json.Marshal(message)
}

func parseAlertRecords(protocolMessages []byte) ([]ProtocolMessage) {
    return []ProtocolMessage{
        &AlertMessage{
            Level: protocolMessages[0],
            Description: protocolMessages[1],
        },
    }
}
