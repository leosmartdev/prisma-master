//Package sar implements specifications for the sarsat beacon identifier.
//The specifications can be found here:
// https://team.technosci.com:8442/browse/CONV-801
package sar

import (
	fmt "fmt"
	"reflect"
	"strconv"
	"testing"
)

func b(s string) byte {
	v, err := strconv.ParseInt(s, 2, 0)
	if err != nil {
		panic(err)
	}
	return byte(v)
}

var bitReadTests = []struct {
	name   string
	b      []byte
	skip   int
	read   int
	result int
}{
	{"one", []byte{b("10000000")}, 0, 1, 1},
	{"two", []byte{b("10000000")}, 0, 2, 2},
	{"skip one", []byte{b("01000000")}, 1, 1, 1},
	{"low bit", []byte{b("00000001")}, 7, 1, 1},
	{"low nibble", []byte{0xab}, 4, 4, 0xb},
	{"high nibble", []byte{0xab}, 0, 4, 0xa},
	{"next low nibble", []byte{0xab, 0xcd}, 12, 4, 0xd},
	{"cross byte", []byte{0xab, 0xcd}, 4, 8, 0xbc},
	{"uint24", []byte{0xab, 0xcd, 0xef}, 0, 24, 0xabcdef},
	{"cross uint24", []byte{0xab, 0xcd, 0xef, 0x98}, 4, 24, 0xbcdef9},
	{"country code 366", []byte{0xad, 0xc6}, 1, 10, 366},
}

func TestReadBits(t *testing.T) {
	for _, test := range bitReadTests {
		t.Run(test.name, func(t *testing.T) {
			bits := NewBitReader(test.b)
			bits.SkipN(test.skip)
			have := bits.ReadN(test.read)
			want := uint64(test.result)
			if want != have {
				t.Errorf("\n want: %v \n have: %v\n", want, have)
			}
		})
	}
}

var bitWriteTests = []struct {
	name   string
	value  int
	skip   int
	writeN int
	result []byte
}{
	{"one", 1, 0, 1, []byte{b("10000000")}},
	{"two", 2, 0, 2, []byte{b("10000000")}},
	{"skip one", 1, 1, 1, []byte{b("01000000")}},
	{"low bit", 1, 0, 8, []byte{b("00000001")}},
	{"low nibble", 0xb, 4, 4, []byte{0x0b}},
	{"high nibble", 0xa, 0, 4, []byte{0xa0}},
	{"next low nibble", 0xd, 12, 4, []byte{0x00, 0x0d}},
	{"cross byte", 0xbc, 4, 8, []byte{0x0b, 0xc0}},
	{"uint24", 0xabcdef, 0, 24, []byte{0xab, 0xcd, 0xef}},
	{"cross uint24", 0xbcdef9, 4, 24, []byte{0x0b, 0xcd, 0xef, 0x90}},
}

func TestWriteBits(t *testing.T) {
	for _, test := range bitWriteTests {
		t.Run(test.name, func(t *testing.T) {
			bits := &BitWriter{}
			bits.WriteZeros(test.skip)
			bits.WriteN(test.writeN, test.value)
			have := bits.bytes
			want := test.result
			if !reflect.DeepEqual(want, have) {
				t.Errorf("\n want: %v \n have: %v\n", want, have)
			}
		})
	}
}

func TestBaudotWrite(t *testing.T) {
	bits := &BitWriter{}
	bits.WriteBaudotN(7, "C7518")
	have := fmt.Sprintf("%08b", bits.bytes[0]) +
		fmt.Sprintf("%08b", bits.bytes[1]) +
		fmt.Sprintf("%08b", bits.bytes[2]) +
		fmt.Sprintf("%08b", bits.bytes[3]) +
		fmt.Sprintf("%08b", bits.bytes[4]) +
		fmt.Sprintf("%02b", bits.bytes[5])
	want := "100100100100101110011100000001011101001100"
	if want != have {
		t.Errorf("\n want: %v \n have: %v\n", want, have)
	}
}

// This is just a double-entry test to make sure cut & paste went to
// the correct characters
var baudotTests = []struct {
	r    rune
	bits string
}{
	{'A', "111000"},
	{'B', "110011"},
	{'C', "101110"},
	{'D', "110010"},
	{'E', "110000"},
	{'F', "110110"},
	{'G', "101011"},
	{'H', "100101"},
	{'I', "101100"},
	{'J', "111010"},
	{'K', "111110"},
	{'L', "101001"},
	{'M', "100111"},
	{'N', "100110"},
	{'O', "100011"},
	{'P', "101101"},
	{'Q', "111101"},
	{'R', "101010"},
	{'S', "110100"},
	{'T', "100001"},
	{'U', "111100"},
	{'V', "101111"},
	{'W', "111001"},
	{'X', "110111"},
	{'Y', "110101"},
	{'Z', "110001"},
	{' ', "100100"},
	{'-', "011000"},
	{'/', "010111"},
	{'0', "001101"},
	{'1', "011101"},
	{'2', "011001"},
	{'3', "010000"},
	{'4', "001010"},
	{'5', "000001"},
	{'6', "010101"},
	{'7', "011100"},
	{'8', "001100"},
	{'9', "000011"},
}

func TestBaudotTable(t *testing.T) {
	for _, test := range baudotTests {
		t.Run(string(test.r), func(t *testing.T) {
			code, ok := runeToBaudot[test.r]
			if !ok {
				t.Errorf("rune not found: %v", test.r)
			}
			want := test.bits
			have := fmt.Sprintf("%06b", code)
			if want != have {
				t.Errorf("\n want: %v \n have: %v\n", want, have)
			}
		})
	}
}
