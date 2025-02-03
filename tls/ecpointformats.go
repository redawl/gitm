package tls

type ECPointFormats struct {
    extension
    Length uint
    ECPointFormatList []ECPointFormat

}

type ECPointFormat struct {
    Type byte
}

func parseECPointFormats(ext extension, extData []byte) Extension {
    ecPointFormats := ECPointFormats{
        extension: ext,
        Length: uint(extData[0]),
    }

    ecPointFormatList := make([]ECPointFormat, ecPointFormats.Length)

    for i := 1; i < len(extData); i++ {
        ecPointFormatList[i-1] = ECPointFormat{
            Type: extData[i],
        }
    }

    ecPointFormats.ECPointFormatList = ecPointFormatList

    return ecPointFormats
}
