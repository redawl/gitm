package tls

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

const (
    CTChangeCipherSpec = 0x14
    CTAlert            = 0x15
    CTHandshake        = 0x16
    CTApplication      = 0x17
    CTHeartbeat        = 0x18
)

const (
    SSL30 = iota
    TLS10
    TLS11
    TLS12
    TLS13
)

func mapCTtoString(ct byte) string {
    switch ct {
    case CTChangeCipherSpec: return "ChangeCipherSpec"
    case CTAlert: return "Alert"
    case CTHandshake: return "Handshake"
    case CTApplication: return "Application"
    case CTHeartbeat: return "HeartBeat"
    default: return "Unknown"
    }
}

func mapVersionToString(major byte, minor byte) string {
    // Commenting out for now, some clients are sending 0x01 instead of 0x03
    // if major != 0x03 {
    //     return fmt.Sprintf("Unknown: %d.%d", major, minor)
    // }

    switch minor {
        case SSL30: return "SSL 3.0"
        case TLS10: return "TLS 1.0"
        case TLS11: return "TLS 1.1"
        case TLS12: return "TLS 1.2"
        case TLS13: return "TLS 1.3"
        default: return fmt.Sprintf("Unkown: %d.%d", major, minor)
    }
}

// TLSRecord This is the general format of all TLS records. 
type TLSRecord struct {
    // This field identifies the Record Layer Protocol Type contained in this record.
    ContentType        byte
    // 
    Version            [2]byte
    // The length of "protocol message(s)", "MAC" and "padding" fields combined (i.e. q−5), not to exceed 214 bytes (16 KiB).
    Length             uint16
    ProtocolMessages   []ProtocolMessage
    // Optional
    MAC                []byte
    // Optional
    Padding            []byte
}

type ProtocolMessage interface {
    GetLogAttrs() []slog.Attr
    MarshalJSON() ([]byte, error)
}

func ParseTLSRecords(message []byte) ([]TLSRecord) {
    length := 0
    records := []TLSRecord{}
    for length < len(message) {
        record := TLSRecord{}
        record.Parse(message[length:])
        length += int(record.Length) + 5
        records = append(records, record)
    }

    return records
}

func (record *TLSRecord) Parse(message []byte) {
    record.ContentType = message[0]
    record.Version = [2]byte{message[2], message[1]}
    record.Length = uint16(message[3]) << 8 + uint16(message[4])
    record.ProtocolMessages = []ProtocolMessage{}

    protocolMessages := message[5:5+int(record.Length)]

    switch record.ContentType {
        case CTHandshake: {
            record.ProtocolMessages = parseHandshakeRecords(protocolMessages)
        }
        case CTAlert: {
            record.ProtocolMessages = parseAlertRecords(protocolMessages)
        }
        case CTChangeCipherSpec: {
            record.ProtocolMessages = parseChangeCipherSpecMessage(protocolMessages)
        }
        case CTApplication: {
            record.ProtocolMessages = parseApplicationMessage(protocolMessages)
            // TODO: Fill out MAC and Padding
        }
        default: {
            record.ProtocolMessages = append(record.ProtocolMessages, &UnknownMessage{
                Tag: "Unknown",
                ProtocolMessages: protocolMessages,
            })
        }
    }
}

func (record TLSRecord) LogAttrs() ([]slog.Attr) {
    attrs := []slog.Attr{
            slog.String("ContentType", mapCTtoString(record.ContentType)),
            slog.String("Version", mapVersionToString(record.Version[1], record.Version[0])),
            slog.Int("Length", int(record.Length)),
    }

    for _, message := range(record.ProtocolMessages) {
        attrs = append(attrs, message.GetLogAttrs()...)
    }

    return attrs
}

func (record TLSRecord) MarshalJSON() ([]byte, error) {
    valueMap := make(map[string]any)

    valueMap["ContentType"] = mapCTtoString(record.ContentType)
    valueMap["Version"] = mapVersionToString(record.Version[1], record.Version[0])
    valueMap["Length"] = int(record.Length)
    valueMap["ProtocolMessages"] = record.ProtocolMessages

    return json.Marshal(valueMap)
}
