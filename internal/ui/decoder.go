package ui

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"iter"
	"maps"
	"strings"
)

var encodingsMap map[string]func(string) (string, error) = make(map[string]func(string)(string, error))

func GetEncodings () iter.Seq[string] {
    return maps.Keys(encodingsMap)
}

func ExecuteEncoding (encoding string, data string) (string, error) {
    encodingFunc := encodingsMap[encoding]

    if encodingFunc == nil {
        return "", fmt.Errorf("Encoding %s is not implemented", encoding)
    }

    return encodingFunc(data)
}

func init() {
    encodingsMap["Url decode"] = url
    encodingsMap["Base64 decode"] = b64
    encodingsMap["Hex decode"] = _hex
}

func url(data string) (string, error) {
    out := strings.Builder{}

    i := 0

    for i < len(data) {
        if data[i] == '%' && i < len(data) - 2 {
            c := make([]byte, 1)
            count, err := hex.Decode(c, []byte(data[i+1:i+3]))

            if err != nil {
                return "", err
            }

            if count != 1 {
                return "", fmt.Errorf("Error decoding bytes: Expected decoding 1 byte, decoded %d bytes", count)
            }
            out.WriteByte(c[0])
            i += 3
        } else { 
            out.WriteByte(data[i])
            i++
        }
    }

    return out.String(), nil
}

func b64 (data string) (string, error) {
    if data != "" {
        reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))

        decoded, err := io.ReadAll(reader)

        if err != nil {
            return "", fmt.Errorf("Error decoding base64: %w", err)
        }

        return string(decoded), nil
    }

    return "", nil
}

func _hex (data string) (string, error) {
    if data != "" {
        reader := hex.NewDecoder(strings.NewReader(data))

        decoded, err := io.ReadAll(reader)

        if err != nil {
            return "", fmt.Errorf("Error decoding hex: %w", err)
        } else {
            return string(decoded), nil
        }
    }

    return "", nil
}
