package nmea

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const (
	// Degrees value
	Degrees = '\u00B0'
	// Minutes value
	Minutes = '\''
	// Seconds value
	Seconds = '"'
	// Point value
	//Point = '.'
	// North value
	North = "N"
	// South value
	South = "S"
	// East value
	East = "E"
	// West value
	West = "W"
)

// LatLong type
type LatLong float64

func DecodeAisChar(character byte) byte {
	character -= 48
	if character > 40 {
		character -= 8
	}
	return character
}

// MessageType returns the type of an AIS message
func MessageType(payload string) uint8 {
	data := []byte(payload[:1])
	return DecodeAisChar(data[0])
}

func BitsToInt(first, last int, payload []byte) uint32 {
	size := uint(last - first) // Bit fields start at 0
	processed, remain := uint(0), uint(0)
	result, temp := uint32(0), uint32(0)

	from := first / 6
	forTimes := last/6 - from

	for i := 0; i <= forTimes; i++ {
		if len(payload) > (from + i) {
			temp = uint32(DecodeAisChar(payload[from+i]))
			if i == 0 {
				remain = uint(first%6) + 1
				processed = 6 - remain
				temp = temp << (31 - processed) >> (31 - size)
			} else if i < forTimes {
				processed = processed + 6
				temp = temp << (size - processed)
			} else {
				remain = uint(last%6) + 1
				temp = temp >> (6 - remain)
			}
			result = result | temp
		}
	}
	return result
}

func BitsToString(first, last int, payload []byte) string {
	length := (last - first + 1) / 6 // How many characters we expect
	start := first / 6               // At which byte the first character starts
	char := uint8(0)

	// Some times we get truncated text fields. Since text fields have constant size,
	// it is frequent that they aren't fully occupied. Transmitters use this to send shorter messages.
	// We should handle this gracefully, adjusting the length of the text we expect to read.
	if len(payload)*6 < last+1 {
		if len(payload)*6 < first+5 { // Haven't seen this case yet (text field missing) but better be prepared
			return ""
		}
		// Do not simplify this. It uses the uint type rounding method to get correct results
		length = (len(payload)*6 - first) / 6
	}

	remain := first % 6

	var text = make([]byte, length) // in order to be able to decode any message size without constraint

	// In this if/else there is some code duplication but I think the speed enhancement is worth it.
	// The other way around would need 2*length branches. Now we have only 2.
	// decodeAisChar function should be safe to use here since we check the payload's length
	if remain < 6 {
		shiftLeftMost := uint8(remain + 2)
		shiftRightMost := uint8(6 - remain)
		for i := 0; i < length; i++ {
			char = DecodeAisChar(payload[start+i])<<shiftLeftMost>>2 |
				DecodeAisChar(payload[start+i+1])>>shiftRightMost
			if char < 32 {
				char += 64
			}
			text[i] = char
		}
	} else {
		for i := 0; i < length; i++ {
			char = DecodeAisChar(payload[start+i])
			if char < 32 {
				char += 64
			}
			text[i] = char
		}
	}

	// We convert to string and trim the righmost spaces and @ according to the format specs.
	//return strings.TrimRight(string(text[:length]), "@ ")
	return strings.Split(string(text[:length]), "@")[0]
}

func CbnBool(bit int, data []byte) bool {
	if BitsToInt(bit, bit, data) == 1 {
		return true
	}
	return false
}

func ParseDecimal(s string) (LatLong, error) {
	// Make sure it parses as a float.
	l, err := strconv.ParseFloat(s, 64)
	if err != nil || s[0] != '-' && len(strings.Split(s, ".")[0]) > 3 {
		return LatLong(0.0), errors.New("parse error (not decimal coordinate)")
	}
	return LatLong(l), nil
}

// NewLatLong parses the supplied string into the LatLong.
//
// Supported formats are:
// - DMS (e.g. 33° 23' 22")
// - Decimal (e.g. 33.23454)
// - GPS (e.g 15113.4322S)
//
func NewLatLong(s string) (LatLong, error) {
	var l LatLong
	var err error
	invalid := LatLong(0.0) // The invalid value to return.
	if l, err = ParseDMS(s); err == nil {
		return l, nil
	} else if l, err = ParseGPS(s); err == nil {
		return l, nil
	} else if l, err = ParseDecimal(s); err == nil {
		return l, nil
	}
	if !l.ValidRange() {
		return invalid, errors.New("coordinate is not in range -180, 180")
	}
	return invalid, fmt.Errorf("cannot parse [%s], unknown format", s)
}

// ParseDMS parses a coordinate in degrees, minutes, seconds.
// - e.g. 33° 23' 22"
func ParseDMS(s string) (LatLong, error) {
	degrees := 0
	minutes := 0
	seconds := 0.0
	// Whether a number has finished parsing (i.e whitespace after it)
	endNumber := false
	// Temporary parse buffer.
	tmpBytes := []byte{}
	var err error

	for i, r := range s {
		if unicode.IsNumber(r) || r == '.' {
			if !endNumber {
				tmpBytes = append(tmpBytes, s[i])
			} else {
				return 0, errors.New("parse error (no delimiter)")
			}
		} else if unicode.IsSpace(r) && len(tmpBytes) > 0 {
			endNumber = true
		} else if r == Degrees {
			if degrees, err = strconv.Atoi(string(tmpBytes)); err != nil {
				return 0, errors.New("parse error (degrees)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		} else if s[i] == Minutes {
			if minutes, err = strconv.Atoi(string(tmpBytes)); err != nil {
				return 0, errors.New("parse error (minutes)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		} else if s[i] == Seconds {
			if seconds, err = strconv.ParseFloat(string(tmpBytes), 64); err != nil {
				return 0, errors.New("parse error (seconds)")
			}
			tmpBytes = tmpBytes[:0]
			endNumber = false
		} else if unicode.IsSpace(r) && len(tmpBytes) == 0 {
			continue
		} else {
			return 0, fmt.Errorf("parse error (unknown symbol [%d])", s[i])
		}
	}
	val := LatLong(float64(degrees) + (float64(minutes) / 60.0) + (float64(seconds) / 60.0 / 60.0))
	return val, nil
}

func ParseGPS(s string) (LatLong, error) {
	parts := strings.Split(s, " ")
	dir := parts[1]
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse error: %s", err.Error())
	}

	degrees := math.Floor(value / 100)
	minutes := value - (degrees * 100)
	value = degrees + minutes/60

	if dir == North || dir == East {
		return LatLong(value), nil
	} else if dir == South || dir == West {
		return LatLong(0 - value), nil
	} else {
		return 0, fmt.Errorf("invalid direction [%s]", dir)
	}
}

func (l LatLong) ValidRange() bool {
	return -180.0 <= l && l <= 180.0
}
