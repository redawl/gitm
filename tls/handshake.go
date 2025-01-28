package tls

import (
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
    MessageType string
    HandshakeMessageDataLength Uint24
    HandshakeMessageData []byte
}

type ClientHello struct {
    LegacyVersion string
    Random                      [32]byte
    LegacySessionId             [32]byte
    CipherSuites                []string
    LegacyCompressionMethods    []string
    Extensions                  []string
}

func (message *HandshakeMessage) GetLogAttrs () []slog.Attr {
    attrs := []slog.Attr{
        slog.String("MessageType", message.MessageType),
        slog.Int("HandshakeMessageDataLength", message.HandshakeMessageDataLength.IntValue()),
        slog.Any("HandshakeMessageData", message.HandshakeMessageData),
    }

    return attrs
}

func parseHandshakeRecords(protocolMessages []byte) ([]ProtocolMessage) {
    messages := []ProtocolMessage{}
    length := 0
    for length < len(protocolMessages) {
        messageType := mapMTtoString(protocolMessages[length])
        switch protocolMessages[length] {
            case MTClientHello: {
                
            }
        }
        messageLength := NewUint24(protocolMessages[length+1], protocolMessages[length+2], protocolMessages[length+3])
        messageData := protocolMessages[length+4:length+4+messageLength.IntValue()]
        messages = append(messages, &HandshakeMessage{
            MessageType: messageType,
            HandshakeMessageDataLength: messageLength,
            HandshakeMessageData: messageData,
        })
        length += 4 + messageLength.IntValue()
    } 

    return messages
}

