//Package sar implements specifications for the sarsat beacon identifier.
//The specifications can be found here:
// https://team.technosci.com:8442/browse/CONV-801
package sar

import (
	fmt "fmt"
	"strings"
)

// Reuse the structure for readers and writers
type bits struct {
	bytes []byte
	pos   int // index into the byte array
	nbits int // bits consumed in the current byte
}

func (b *bits) String() string {
	fbytes := make([]string, len(b.bytes))
	for i, it := range b.bytes {
		fbytes[i] = fmt.Sprintf("%08b", it)
	}
	return strings.Join(fbytes, " ")
}

// BitReader reads bits from a byte array. Reads start at the first byte
// in the array at the MSB and works towards the LSB.
type BitReader struct {
	bits
}

// NewBitReader creates a new reader for the given series of bytes. The value
// passed in will be mutated, make a copy if necessary.
func NewBitReader(data []byte) *BitReader {
	b := &BitReader{
		bits: bits{
			bytes: data,
		},
	}
	return b
}

func (b *BitReader) Read() uint64 {
	v := b.bytes[b.pos] & 0x80
	b.bytes[b.pos] <<= 1
	b.nbits++
	if b.nbits == 8 {
		b.nbits = 0
		b.pos++
	}
	if v != 0 {
		return 1
	}
	return 0
}

func (b *BitReader) ReadN(n int) uint64 {
	var v uint64
	for i := 0; i < n; i++ {
		v <<= 1
		v |= b.Read()
	}
	return v
}

func (b *BitReader) SkipN(n int) {
	_ = b.ReadN(n)
}

func (b *BitReader) ReadBaudot() (rune, error) {
	return b.readBaudot(baudotSixBit)
}

func (b *BitReader) ReadBaudot5() (rune, error) {
	return b.readBaudot(baudotFiveBit)
}

func (b *BitReader) readBaudot(size baudotSize) (rune, error) {
	var code int
	switch size {
	case baudotSixBit:
		code = int(b.ReadN(6))
	case baudotFiveBit:
		code = int(b.ReadN(5)) | 32 // always set 6th bit 0b100000
	default:
		return 0, fmt.Errorf("invalid size: %v", size)
	}
	ch, ok := baudotToRune[code]
	if !ok {
		return 0, fmt.Errorf("invalid baudot code: %v", code)
	}
	return ch, nil
}

func (b *BitReader) ReadBaudotN(n int) (string, error) {
	return b.readBaudotN(n, baudotSixBit)
}

func (b *BitReader) ReadBaudot5N(n int) (string, error) {
	return b.readBaudotN(n, baudotFiveBit)
}

func (b *BitReader) readBaudotN(n int, size baudotSize) (string, error) {
	runes := make([]rune, n)
	for i := 0; i < n; i++ {
		ch, err := b.readBaudot(size)
		if err != nil {
			return "", err
		}
		runes[i] = ch
	}
	return string(runes), nil
}

// BitWriter writes bits to a byte array. Writes start at the first byte
// in the array at the MSB and works towards the LSB.

type BitWriter struct {
	bits
}

func (b *BitWriter) Write(val int) *BitWriter {
	if b.bytes == nil {
		b.bytes = make([]byte, 1)
		b.nbits = 8
	} else if b.nbits == 0 {
		b.bytes = append(b.bytes, byte(0))
		b.nbits = 8
		b.pos++
	}

	b.nbits--
	if val != 0 {
		val = 1 << uint(b.nbits)
	}
	b.bytes[b.pos] |= byte(val)
	return b
}

func (b *BitWriter) WriteN(n int, val int) *BitWriter {
	for i := n - 1; i >= 0; i-- {
		mask := 1 << uint(i)
		b.Write(val & mask)
	}
	return b
}

func (b *BitWriter) WriteZeros(n int) *BitWriter {
	for i := 0; i < n; i++ {
		b.Write(0)
	}
	return b
}

func (b *BitWriter) WriteBaudot(r rune) error {
	return b.writeBaudot(r, baudotSixBit)
}

func (b *BitWriter) WriteBaudot5(r rune) error {
	return b.writeBaudot(r, baudotFiveBit)
}

func (b *BitWriter) writeBaudot(r rune, size baudotSize) error {
	code, ok := runeToBaudot[r]
	if !ok {
		return fmt.Errorf("no encoding for rune: %v", r)
	}
	switch size {
	case baudotSixBit:
		b.WriteN(6, code)
	case baudotFiveBit:
		b.WriteN(5, code)
	default:
		return fmt.Errorf("invalid size: %v", size)
	}
	return nil
}

func (b *BitWriter) WriteBaudotN(n int, str string) error {
	return b.writeBaudotN(n, str, baudotSixBit)
}

func (b *BitWriter) WriteBaudot5N(n int, str string) error {
	return b.writeBaudotN(n, str, baudotFiveBit)
}

func (b *BitWriter) writeBaudotN(n int, str string, size baudotSize) error {
	pad := n - len(str)
	for i := 0; i < pad; i++ {
		if err := b.writeBaudot(' ', size); err != nil {
			return err
		}
	}
	for _, ch := range str {
		if err := b.writeBaudot(rune(ch), size); err != nil {
			return err
		}
	}
	return nil
}

// Modified-Baudot is normally six bits but the serial standard location
// protocol message with a aircraft operator desginator uses a five bit
// variant
type baudotSize bool

const (
	baudotSixBit  baudotSize = true
	baudotFiveBit            = false
)

var baudotToRune = map[int]rune{
	56: 'A', // 0b111000
	51: 'B', // 0b110011
	46: 'C', // 0b101110
	50: 'D', // 0b110010
	48: 'E', // 0b110000
	54: 'F', // 0b110110
	43: 'G', // 0b101011
	37: 'H', // 0b100101
	44: 'I', // 0b101100
	58: 'J', // 0b111010
	62: 'K', // 0b111110
	41: 'L', // 0b101001
	39: 'M', // 0b100111
	38: 'N', // 0b100110
	35: 'O', // 0b100011
	45: 'P', // 0b101101
	61: 'Q', // 0b111101
	42: 'R', // 0b101010
	52: 'S', // 0b110100
	33: 'T', // 0b100001
	60: 'U', // 0b111100
	47: 'V', // 0b101111
	57: 'W', // 0b111001
	55: 'X', // 0b110111
	53: 'Y', // 0b110101
	49: 'Z', // 0b110001
	36: ' ', // 0b100100
	24: '-', // 0b011000
	23: '/', // 0b010111
	13: '0', // 0b001101
	29: '1', // 0b011101
	25: '2', // 0b011001
	16: '3', // 0b010000
	10: '4', // 0b001010
	1:  '5', // 0b000001
	21: '6', // 0b010101
	28: '7', // 0b011100
	12: '8', // 0b001100
	3:  '9', // 0b000011
}

var runeToBaudot = map[rune]int{}

func init() {
	for code, r := range baudotToRune {
		runeToBaudot[r] = code
	}
}
