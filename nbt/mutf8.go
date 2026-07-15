package nbt

import (
	"errors"
	"fmt"
	"unicode/utf16"
)

func mutf8Len(s string) int {
	n := 0
	for _, r := range s {
		switch {
		case r >= 0x0001 && r <= 0x007F:
			n++
		case r == 0x0000 || r <= 0x07FF:
			n += 2
		case r <= 0xFFFF:
			n += 3
		default:
			n += 6
		}
	}

	return n
}

func encodeMUTF8(dst []byte, s string) []byte {
	for _, r := range s {
		if r > 0xFFFF {
			high, low := utf16.EncodeRune(r)
			dst = encodeUnit(dst, uint16(high))
			dst = encodeUnit(dst, uint16(low))
			continue
		}

		dst = encodeUnit(dst, uint16(r))
	}

	return dst
}

func encodeUnit(dst []byte, unit uint16) []byte {
	switch {
	case unit >= 0x0001 && unit <= 0x007F:
		return append(dst, byte(unit))
	case unit == 0x0000 || unit <= 0x07FF:
		return append(dst, 0xC0|byte(unit>>6), 0x80|byte(unit&0x3F))
	default:
		return append(dst, 0xE0|byte(unit>>12), 0x80|byte((unit>>6)&0x3F), 0x80|byte(unit&0x3F))
	}
}

func decodeMUTF8(raw []byte) (string, error) {
	units := make([]uint16, 0, len(raw))

	for i := 0; i < len(raw); {
		lead := raw[i]

		switch {
		case lead&0x80 == 0x00:
			units = append(units, uint16(lead))
			i++
		case lead&0xE0 == 0xC0:
			if i+1 >= len(raw) {
				return "", errors.New("nbt: truncated 2-byte mutf-8 sequence")
			}
			if raw[i+1]&0xC0 != 0x80 {
				return "", fmt.Errorf("nbt: invalid mutf-8 continuation byte %#x", raw[i+1])
			}
			unit := uint16(lead&0x1F)<<6 | uint16(raw[i+1]&0x3F)
			if unit != 0 && unit < 0x80 {
				return "", fmt.Errorf("nbt: overlong mutf-8 encoding of %#x", unit)
			}
			units = append(units, unit)
			i += 2
		case lead&0xF0 == 0xE0:
			if i+2 >= len(raw) {
				return "", errors.New("nbt: truncated 3-byte mutf-8 sequence")
			}
			if raw[i+1]&0xC0 != 0x80 || raw[i+2]&0xC0 != 0x80 {
				return "", errors.New("nbt: invalid mutf-8 continuation byte")
			}
			unit := uint16(lead&0x0F)<<12 | uint16(raw[i+1]&0x3F)<<6 | uint16(raw[i+2]&0x3F)
			if unit < 0x800 {
				return "", fmt.Errorf("nbt: overlong mutf-8 encoding of %#x", unit)
			}
			units = append(units, unit)
			i += 3
		default:
			return "", fmt.Errorf("nbt: invalid mutf-8 lead byte %#x", lead)
		}
	}

	return string(utf16.Decode(units)), nil
}
