package tls

import "log/slog"

type ServerHello struct {
    HandshakeMessage
    LegacyVersion string
    Random        [32]byte
    // LegacySessionIdEcho length is 0 - 32 bytes
    LegacySessionIdEcho []byte
    CipherSuite         [2]byte
    LegacyCompressionMethod byte
    Extensions []Extension
}

func (serverHello ServerHello) GetLogAttrs () []slog.Attr {
    attrs := serverHello.HandshakeMessage.GetLogAttrs()
    additionalAttrs := []slog.Attr{
        slog.String("LegacyVersion", serverHello.LegacyVersion),
        slog.Any("Random", serverHello.Random),
        slog.Any("LegacySessionIdEcho", serverHello.LegacySessionIdEcho),
        slog.Any("CipherSuite", serverHello.CipherSuite),
        slog.Any("LegacyCompressionMethod", serverHello.LegacyCompressionMethod),
    }
    attrs = append(attrs[:len(attrs) - 2], additionalAttrs...)
    for _, extension := range(serverHello.Extensions) {
        attrs = append(attrs, extension.GetLogAttr())
    }

    return attrs
}

func parseServerHelloMessage(handshake *HandshakeMessage, messageData []byte) ServerHello {
    i := 0
    legacyVersion := mapVersionToString(messageData[i+1], messageData[i])
    i += 2
    random := messageData[i:i+32]
    i += 32
    sessionIdLength := int(messageData[i])
    i += 1
    legacySessionIdEcho := messageData[i:i+sessionIdLength]
    i += sessionIdLength
    cipherSuite := [2]byte{messageData[i], messageData[i + 1]}
    i += 2
    legacyCompressionMethod := messageData[i]
    i += 1
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

    return ServerHello{
        HandshakeMessage: *handshake,
        LegacyVersion: legacyVersion,
        Random: [32]byte(random),
        LegacySessionIdEcho: legacySessionIdEcho,
        LegacyCompressionMethod: legacyCompressionMethod,
        CipherSuite: cipherSuite,
        Extensions: extensions,
    }
}
