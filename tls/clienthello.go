package tls

import (
	"encoding/json"
	"log/slog"
)

type ClientHello struct {
    HandshakeMessage
    LegacyVersion [2]byte
    Random                      [32]byte
    // LegacySessionId length 0-32 bytes
    LegacySessionId             []byte
    CipherSuites                [][2]byte
    LegacyCompressionMethods    []byte
    Extensions                  []Extension
}

func (clientHello ClientHello) GetLogAttrs () []slog.Attr {
    attrs := clientHello.HandshakeMessage.GetLogAttrs()
    additionalAttrs := []slog.Attr{
        slog.String("LegacyVersion", mapVersionToString(clientHello.LegacyVersion[1], clientHello.LegacyVersion[0])),
        slog.Any("Random", clientHello.Random),
        slog.Any("LegacySessionId", clientHello.LegacySessionId),
        slog.Any("CipherSuites", clientHello.CipherSuites),
        slog.Any("LegacyCompressionMethods", clientHello.LegacyCompressionMethods),
    }
    attrs = append(attrs[:len(attrs) - 2], additionalAttrs...)

    for _, extension := range(clientHello.Extensions) {
        attrs = append(attrs, extension.GetLogAttr())
    }

    return attrs
}

func (clientHello ClientHello) MarshalJSON() ([]byte, error) {
    valueMap := clientHello.HandshakeMessage.getValueMap()
    valueMap["LegacyVersion"] = mapVersionToString(clientHello.LegacyVersion[1], clientHello.LegacyVersion[0])
    valueMap["Random"] = clientHello.Random
    valueMap["LegacySessionId"] = clientHello.LegacySessionId
    valueMap["CipherSuites"] = clientHello.CipherSuites
    valueMap["LegacyCompressionMethods"] = clientHello.LegacyCompressionMethods
    valueMap["Extensions"] = clientHello.Extensions

    return json.Marshal(valueMap)
}

func parseClientHelloMessage(handshake *HandshakeMessage, messageData []byte) ClientHello {
    i := 0
    legacyVersion := [2]byte{messageData[i+1], messageData[i]}
    i += 2
    random := messageData[i:i+32]
    i += 32
    sessionIdLength := int(messageData[i])
    i += 1
    legacySessionId := messageData[i:i+sessionIdLength]
    i += sessionIdLength
    cipherSuitesLength := int(messageData[i]) << 8 + int(messageData[i+1])
    i += 2
    cipherSuites := [][2]byte{}
    for j := i; j < i + cipherSuitesLength; j += 2 {
        cipherSuites = append(cipherSuites, [2]byte{messageData[j], messageData[j+1]})
    }
    i += cipherSuitesLength
    compressionMethodsLength := int(messageData[i])
    i += 1
    compressionMethods := []byte{}
    for j := i; j < i + compressionMethodsLength; j++ {
        compressionMethods = append(compressionMethods, messageData[j])
    }
    i += compressionMethodsLength

    extensionsLength := int(messageData[i]) << 8 + int(messageData[i+1])
    i += 2
    extensions := []Extension{}

    for j := i; j < i + extensionsLength; {
        extensionType := messageData[j:j+2]
        extensionLength := int(messageData[j+2]) << 8 + int(messageData[j+3])
        extensions = append(extensions, CreateExtension(
            [2]byte{extensionType[0], extensionType[1]},
            byte(extensionLength),
            messageData[j+4:j+4+extensionLength],
        ))

        j += 4 + extensionLength
    }

    return ClientHello{
        HandshakeMessage: *handshake,
        LegacyVersion: legacyVersion,
        Random: [32]byte(random),
        LegacySessionId: legacySessionId,
        CipherSuites: cipherSuites,
        LegacyCompressionMethods: compressionMethods,
        Extensions: extensions,
    }
}
