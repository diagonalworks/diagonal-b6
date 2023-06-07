package b6

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	gbPostcodeElementBits = 6 // to encode a-z and 0-9
	gbPostcodeMinLength   = 5
	gbPostcodeMaxLength   = 7
	gbPostcodeLengthBits  = 2 // to encode a maximum of (7-5) = 2
)

func PointIDFromGBPostcode(postcode string) PointID {
	postcode = strings.ToUpper(strings.Replace(postcode, " ", "", -1))
	if len(postcode) < gbPostcodeMinLength || len(postcode) > gbPostcodeMaxLength {
		return PointIDInvalid
	}
	id := uint64(0)
	const shift = 6 // 6 bits
	for i, r := range postcode {
		if i > 0 {
			id <<= gbPostcodeElementBits
		}
		var v uint64
		if r >= '0' && r <= '9' {
			v = uint64(r - '0')
		} else if r >= 'A' && r <= 'Z' {
			v = uint64(r-'A') + 10
		} else {
			return PointIDInvalid
		}
		id |= v
	}
	id <<= gbPostcodeLengthBits
	id |= uint64(len(postcode) - gbPostcodeMinLength)
	return MakePointID(NamespaceGBCodePoint, id)
}

func PostcodeFromPointID(id PointID) (string, bool) {
	if id.Namespace != NamespaceGBCodePoint {
		return "", false
	}
	postcode := ""
	n := gbPostcodeMinLength + int(id.Value&((1<<gbPostcodeLengthBits)-1))
	id.Value >>= gbPostcodeLengthBits
	for i := 0; i < n; i++ {
		v := id.Value & ((1 << gbPostcodeElementBits) - 1)
		if v >= 0 && v < 10 {
			postcode = string('0'+rune(v)) + postcode
		} else if v >= 10 && v < 10+26 {
			postcode = string('A'+rune(v-10)) + postcode
		} else {
			return "", false
		}
		id.Value >>= gbPostcodeElementBits
	}
	return postcode, true
}

const (
	ukONSCodeShift  = 40
	ukONSYearShift  = 32
	ukONSYearMask   = 0xff
	ukONSLetterMask = 0xff
	ukONSNumberMask = 0xffffffff
)

func FeatureIDFromGBONSCode(code string, year int, t FeatureType) FeatureID {
	// ONS codes are a letter followed by 8 digits
	if len(code) != 9 {
		return FeatureIDInvalid
	}
	n, err := strconv.Atoi(code[1:])
	if err != nil {
		return FeatureIDInvalid
	}
	codeBits := uint64(uint8(byte(code[0]))) << ukONSCodeShift
	yearBits := uint64(uint8(year-1900)) << ukONSYearShift
	return FeatureID{Type: t, Namespace: NamespaceUKONSBoundaries, Value: codeBits | yearBits | uint64(n)}
}

func GBONSCodeFromFeatureID(id FeatureID) (string, int, bool) {
	if id.Namespace != NamespaceUKONSBoundaries {
		return "", 0, false
	}
	year := int((id.Value>>ukONSYearShift)&ukONSYearMask) + 1900
	letter := string(byte((id.Value >> ukONSCodeShift) & ukONSLetterMask))
	return fmt.Sprintf("%s%08d", letter, id.Value&ukONSNumberMask), year, true
}
