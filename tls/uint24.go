package tls

type Uint24 struct {
    value [3]uint8
}

func NewUint24(b1 byte, b2 byte, b3 byte) (Uint24) {
    uint24 := Uint24{
        value: [3]uint8{uint8(b1), uint8(b2), uint8(b3)},
    }

    return uint24
}

func (u *Uint24) IntValue() int {
    return int(u.value[0]) << 16 + int(u.value[1]) << 8 + int(u.value[2])
}
