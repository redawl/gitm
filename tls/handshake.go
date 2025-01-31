package tls

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

const (
    MTHelloRequest        = 0
    MTClientHello         = 1
    MTServerHello         = 2
    MTNewSessionTicket    = 4
    MTEncryptedExtensions = 8
    MTCertificate         = 11
    MTServerKeyExchange   = 12
    MTCertificateRequest  = 13
    MTServerHelloDone     = 14
    MTCertificateVerify   = 15
    MTClientKeyExchange   = 16
    MTFinished            = 20
)

func mapMTtoString(mt byte) string {
    switch mt {
    case MTHelloRequest: return "HelloRequest"
    case MTClientHello: return "ClientHello"
    case MTServerHello: return "ServerHello"
    case MTNewSessionTicket: return "NewSessionTicket"
    case MTEncryptedExtensions: return "Encrypted Extensions"
    case MTCertificate: return "Certificate"
    case MTServerKeyExchange: return "ServerKeyExchange"
    case MTCertificateRequest: return "Certificate Request"
    case MTServerHelloDone: return "ServerHelloDone"
    case MTCertificateVerify: return "CertificateVerify"
    case MTClientKeyExchange: return "ClientKeyExchange"
    case MTFinished: return "Finished"
    default: return fmt.Sprintf("Unknown: %d", mt)
    }
}

type HandshakeMessage struct {
    MessageType byte
    HandshakeMessageDataLength Uint24
}

func (message HandshakeMessage) GetLogAttrs () []slog.Attr {
    attrs := []slog.Attr{
        slog.String("MessageType", mapMTtoString(message.MessageType)),
        slog.Int("HandshakeMessageDataLength", message.HandshakeMessageDataLength.IntValue()),
    }

    return attrs
}

func (message HandshakeMessage) getValueMap() map[string]any {
    valueMap := make(map[string]any)
    valueMap["MessageType"] = mapMTtoString(message.MessageType)
    valueMap["HandshakeMessageDataLength"] = message.HandshakeMessageDataLength.IntValue()

    return valueMap
}

func (message HandshakeMessage) MarshalJSON() ([]byte, error) {
    return json.Marshal(message.getValueMap())
}

func parseHandshakeRecords(protocolMessages []byte) ([]ProtocolMessage) {
    messages := []ProtocolMessage{}
    length := 0
    for length < len(protocolMessages) {
        messageType := protocolMessages[length]
        messageLength := NewUint24(protocolMessages[length+1], protocolMessages[length+2], protocolMessages[length+3])

        handshakeRecord := &HandshakeMessage{
            MessageType: messageType,
            HandshakeMessageDataLength: messageLength,
        }

        var protocolMessage ProtocolMessage

        switch messageType {
            case MTClientHello: {
                protocolMessage = parseClientHelloMessage(handshakeRecord, protocolMessages[length+4:length+4+messageLength.IntValue()])
            }
            case MTServerHello: {
                protocolMessage = parseServerHelloMessage(handshakeRecord, protocolMessages[length+4:length+4+messageLength.IntValue()])
            }
            default: {
                protocolMessage = handshakeRecord
            }
        }

        messages = append(messages, protocolMessage)
        length += 4 + messageLength.IntValue()
    } 

    return messages
}

