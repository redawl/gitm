package tls

import (
	"encoding/json"
	"log/slog"
)

const (
    EXT_SERVER_NAME                            = 0
    EXT_MAX_FRAGMENT_LENGTH                    = 1
    EXT_STATUS_REQUEST                         = 5
    EXT_SUPPORTED_GROUPS                       = 10
    EXT_EC_POINT_FORMATS                       = 11
    EXT_SIGNATURE_ALGORITHMS                   = 13
    EXT_USE_SRTP                               = 14
    EXT_HEARTBEAT                              = 15
    EXT_APPLICATION_LAYER_PROTOCOL_NEGOTIATION = 16
    EXT_SIGNED_CERTIFICATE_TIMESTAMP           = 18
    EXT_CLIENT_CERTIFICATE_TYPE                = 19
    EXT_SERVER_CERTIFICATE_TYPE                = 20
    EXT_PADDING                                = 21
    EXT_ENCRYPT_THEN_MAC                       = 22
    EXT_EXTEND_MASTER_SECRET                   = 23
    EXT_PRE_SHARED_KEY                         = 41
    EXT_EARLY_DATA                             = 42
    EXT_SUPPORTED_VERSIONS                     = 43
    EXT_COOKIE                                 = 44
    EXT_PSK_KEY_EXCHANGE_MODES                 = 45
    EXT_CERTIFICATE_AUTHORITIES                = 47
    EXT_OID_FILTERS                            = 48
    EXT_POST_HANDSHAKE_AUTH                    = 49
    EXT_SIGNATURE_ALGORITHMS_CERT              = 50
    EXT_KEY_SHARE                              = 51
)

func mapTypeToString(extType [2]byte) string {
    switch t := int(extType[0]) << 8 + int(extType[1]); t {
        case EXT_SERVER_NAME:
            return "ServerName"
        case EXT_MAX_FRAGMENT_LENGTH:
            return "MaxFragmentLength"
        case EXT_STATUS_REQUEST:
            return "StatusRequest"
        case EXT_SUPPORTED_GROUPS:
            return "SupportedGroups"
        case EXT_EC_POINT_FORMATS:
            return "ECPointFormats"
        case EXT_SIGNATURE_ALGORITHMS:
            return "SignatureAlgorithms"
        case EXT_USE_SRTP:
            return "UseSRTP"
        case EXT_HEARTBEAT:
            return "Heartbeat"
        case EXT_APPLICATION_LAYER_PROTOCOL_NEGOTIATION:
            return "ApplicationLayerProtocolNegotiation"
        case EXT_SIGNED_CERTIFICATE_TIMESTAMP:
            return "SignedCertificateTimestamp"
        case EXT_CLIENT_CERTIFICATE_TYPE:
            return "ClientCertificateType"
        case EXT_SERVER_CERTIFICATE_TYPE:
            return "ServerCertificateType"
        case EXT_PADDING:
            return "Padding"
        case EXT_ENCRYPT_THEN_MAC:
            return "EncryptThenMAC"
        case EXT_EXTEND_MASTER_SECRET:
            return "ExtendMasterSecret"
        case EXT_PRE_SHARED_KEY:
            return "PreSharedKey"
        case EXT_EARLY_DATA:
            return "EarlyData"
        case EXT_SUPPORTED_VERSIONS:
            return "SupportedVersions"
        case EXT_COOKIE:
            return "Cookie"
        case EXT_PSK_KEY_EXCHANGE_MODES:
            return "PskKeyExchangeModes"
        case EXT_CERTIFICATE_AUTHORITIES:
            return "CertificateAuthorities"
        case EXT_OID_FILTERS:
            return "OIDFilters"
        case EXT_POST_HANDSHAKE_AUTH:
            return "PostHandshakeAuth"
        case EXT_SIGNATURE_ALGORITHMS_CERT:
            return "SignatureAlgorithmsCert"
        case EXT_KEY_SHARE:
            return "KeyShare"
        default:
            return "Unknown"
    }
}

type Extension interface {
    GetLogAttr() slog.Attr
    MarshalJSON() ([]byte, error)
}

type extension struct {
    Type [2]byte
    Length byte
}

func (ext extension) GetLogAttr() slog.Attr {
    return slog.Group("Extension", 
        slog.String("type", mapTypeToString(ext.Type)),
        slog.Any("length", ext.Length),
    )
}

func (ext extension) getValueMap () map[string]any {
    valueMap := make(map[string]any)
    valueMap["type"] = mapTypeToString(ext.Type)
    valueMap["length"] = int(ext.Length)

    return valueMap
}

func (ext extension) MarshalJSON() ([]byte, error) {
    return json.Marshal(ext.getValueMap())
}

func CreateExtension(extType [2]byte, extLength byte, extData []byte) Extension {
    ext := extension{
        Type: extType,
        Length: extLength,
    }

    switch t := int(extType[0]) << 8 + int(extType[1]); t {
        case EXT_SERVER_NAME: {
            return parseServerNameIndication(ext, extData)
        }
        // case EXT_MAX_FRAGMENT_LENGTH: {
        // return parseMaximumFragmentLength(ext, extData)
        // }
        case EXT_EC_POINT_FORMATS: {
            return parseECPointFormats(ext, extData)
        }
        default: 
            return extension{
                Type: extType,
                Length: extLength,
            }
    }
}
