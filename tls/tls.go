package tls

const (
    CTChangeCipherSpec = 0x14
    CTAlert            = 0x15
    CTHandshake        = 0x16
    CTApplication      = 0x17
    CTHeartbeat        = 0x18
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

// TLSRecord This is the general format of all TLS records. 
type TLSRecord struct {
    // This field identifies the Record Layer Protocol Type contained in this record.
    ContentType        string
    // 
    LegacyVersionMajor byte
    LegacyVersionMinor byte
    // The length of "protocol message(s)", "MAC" and "padding" fields combined (i.e. q−5), not to exceed 214 bytes (16 KiB).
    Length             uint16
    ProtocolMessages   []byte

}

type HandshakeMessage struct {
    MessageDataLength []byte
    HandshakeMessageData []byte
}

type TLSHandshakeRecord struct {
    TLSRecord
    MessageType byte
    HandshakeMessages []HandshakeMessage
}

func ParseTLSRecords(message []byte) ([]*TLSRecord) {
    length := 0
    records := []*TLSRecord{}
    for length < len(message) {
        record := ParseTLSRecord(message[length:])
        length += int(record.Length) + 5
        records = append(records, record)
    }

    return records
}

func ParseTLSRecord(message []byte) (*TLSRecord) {
    record := &TLSRecord{
        ContentType: mapCTtoString(message[0]),
        LegacyVersionMajor: message[1],
        LegacyVersionMinor: message[2],
        Length: uint16(message[3]) << 8 + uint16(message[4]),
    }

    record.ProtocolMessages = message[5:5+int(record.Length)]

    return record
}
