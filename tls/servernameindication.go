package tls

import "encoding/json"

const (
    SNI_TYPE_HOSTNAME = 0x0000
)

func mapSNITYPEToString (sniType byte) string {
    switch sniType {
        case SNI_TYPE_HOSTNAME: 
            return "Hostname"
        default:
            return "Unknown"
    }
}

type ServerNameIndication struct {
    extension
    Length uint16
    ServerNameList []ServerName
}

type ServerName struct {
    Type byte
    Length uint16
    ServerName string
}

func (sni ServerNameIndication) MarshalJSON() ([]byte, error) {
    valueMap := sni.extension.getValueMap()
    valueMap["SNILength"] = sni.Length
    
    snList := []map[string]any{}

    for _, sn := range(sni.ServerNameList) {
        snMap := make(map[string]any)
        snMap["ServerNameType"] = mapSNITYPEToString(sn.Type)
        snMap["ServerNameLength"] = sn.Length
        snMap["ServerName"] = sn.ServerName
        snList = append(snList, snMap)
    }

    valueMap["ServerNameList"] = snList

    return json.Marshal(valueMap)
}

func parseServerNameIndication (ext extension, extData []byte) ServerNameIndication {
    length := uint16(extData[0]) << 8 + uint16(extData[1])
    sniList := []ServerName{} 
    for i := 2; i < int(length); {
        sniType := extData[i]
        sniLength := uint16(extData[i+1]) << 8 +  uint16(extData[i+2])
        sni := extData[i+3:i+3+int(sniLength)]
        sniList = append(sniList, ServerName{
            Type: sniType,
            Length: sniLength,
            ServerName: string(sni),
        })
        i += 3 + int(sniLength)
    }

    return ServerNameIndication{
        extension: ext,
        Length: length,
        ServerNameList: sniList,
    }
}
