package encoders

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/bits"
	"strings"
)

type BitEncoder interface {
	EncodeBytes([]byte) []byte
	DecodeBytes([]byte) ([]byte, error)
	EncodedLen([]byte) int
	DecodedLen([]byte) int
	Name() string
}

var (
	AllEncoders = []BitEncoder{&encoder8b10b{}}
)

func init() {
	names := make([]string, len(AllEncoders))
	for i, v := range AllEncoders {
		names[i] = v.Name()
	}
	log.Printf("Registered encoders: %s", strings.Join(names, ", "))
}

type encoder8b10b struct {
	rd int
}

func NewEncoder8B10B() BitEncoder {
	return &encoder8b10b{}
}

func (enc *encoder8b10b) Name() string {
	return "8b10b encoder"
}

// Note that if the input is not a multiple of 4 bytes, this can end up with
// extra 0 byte values at the end
func (enc *encoder8b10b) EncodeBytes(src []byte) []byte {
	var buf []byte
	enc.rd = -1
	for i := 0; i < len(src); i += 4 {
		end := i + 4
		if end > len(src) {
			end = len(src)
		}
		piece := src[i:end]
		encoded := enc.encodePiece(piece)
		endpiece := make([]byte, 8)
		binary.BigEndian.PutUint64(endpiece, encoded)
		buf = append(buf, endpiece[3:]...)
	}
	return buf
}

func (enc *encoder8b10b) DecodeBytes(src []byte) ([]byte, error) {
	var buf []byte
	for i := 0; i < len(src); i += 5 {
		end := i + 5
		if end > len(src) {
			end = len(src)
		}
		piece := src[i:end]
		decoded, err := enc.decodePiece(piece)
		if err != nil {
			return nil, err
		}
		buf = append(buf, decoded...)
	}
	return buf, nil
}

func (enc *encoder8b10b) EncodedLen(v []byte) int {
	return len(v) * 10 / 8
}

func (enc *encoder8b10b) DecodedLen(v []byte) int {
	return len(v) * 8 / 10
}

// decode 40 bits to 32 bits
func (enc *encoder8b10b) decodePiece(piece []byte) ([]byte, error) {
	buf := make([]byte, 8)
	for i, v := range piece {
		buf[i+3] = v
	}
	val := binary.BigEndian.Uint64(buf)
	buf = buf[:0]
	for i := 0; i < 4; i++ {
		chunk := (val >> ((3 - i) * 10)) & 0x3ff
		if chunk == 0 {
			// special case, we treat as EOF
			break
		}
		dec, err := enc.decodeByte(chunk)
		if err != nil {
			return nil, err
		}
		buf = append(buf, dec)
	}
	return buf, nil
}

// decode lowest 10 bits to 8 bits
func (enc *encoder8b10b) decodeByte(val uint64) (byte, error) {
	bigWord := int(val >> 4)
	littleWord := int(val & 0xf)
	var rv byte
	if v := table4b3b[littleWord]; v.valid {
		rv |= (v.val << 5)
	} else {
		return 0, fmt.Errorf("unable to decode 4b value %x", littleWord)
	}
	if v := table6b5b[bigWord]; v.valid {
		rv |= v.val
	} else {
		return 0, fmt.Errorf("unable to decode 6b value %x", bigWord)
	}
	return rv, nil
}

// encodes 32 bits into the *bottom 40* bits of a uint64
// less than 32 bits still encodes into the same place with the lower bits
// being 0
func (enc *encoder8b10b) encodePiece(val []byte) uint64 {
	if len(val) > 4 {
		panic("error in encodePiece: too much data")
	}
	var res uint64
	for i, v := range val {
		e := enc.encodeByte(v)
		res |= (e << ((3 - i) * 10))
	}
	return res
}

// encode a single byte to the bottom 10 bits
func (enc *encoder8b10b) encodeByte(v byte) uint64 {
	lbits := v & 0x1f
	hbits := (v >> 5) & 0x7

	// now lookup the low bits, updating rd
	row := table5b6b[lbits]
	var enc6b byte
	if len(row) > 1 {
		if enc.rd == -1 {
			enc6b = row[0]
		} else {
			enc6b = row[1]
		}
	} else {
		enc6b = row[0]
	}
	pc := bits.OnesCount8(enc6b)
	if pc == 4 {
		enc.rd = 1
	} else if pc == 2 {
		enc.rd = -1
	}
	rv := uint64(enc6b) << 4

	// now do the high bits, special casing 7
	var enc4b byte
	row = table3b4b[hbits]
	if hbits == 7 {
		if (enc.rd == -1 && (lbits == 17 || lbits == 18 || lbits == 20)) ||
			(enc.rd == 1 && (lbits == 11 || lbits == 13 || lbits == 14)) {
			if enc.rd == -1 {
				enc4b = row[2]
			} else {
				enc4b = row[3]
			}
		} else {
			if enc.rd == -1 {
				enc4b = row[0]
			} else {
				enc4b = row[1]
			}
		}
	} else {
		if len(row) > 1 {
			if enc.rd == -1 {
				enc4b = row[0]
			} else {
				enc4b = row[1]
			}
		} else {
			enc4b = row[0]
		}
	}
	pc = bits.OnesCount8(enc4b)
	if pc == 3 {
		enc.rd = 1
	} else if pc == 1 {
		enc.rd = -1
	}
	rv |= uint64(enc4b)
	log.Printf("enc byte 0x%02x to 0x%03x", v, rv)
	return rv
}

func reverse6bits(b byte) byte {
	return table6brev[int(b&0x3f)]
}

func reverse4bits(b byte) byte {
	return table4brev[int(b&0xf)]
}

func init() {
	table6brev = make([]byte, 64)
	for i := 0; i < 64; i++ {
		k := (((i & 0x1) << 5) |
			((i & 0x2) << 3) |
			((i & 0x4) << 1) |
			((i & 0x8) >> 1) |
			((i & 0x10) >> 3) |
			((i & 0x20) >> 5))
		table6brev[i] = byte(k)
	}
	table4brev = make([]byte, 16)
	for i := 0; i < 16; i++ {
		k := (((i & 0x1) << 3) |
			((i & 0x2) << 1) |
			((i & 0x4) >> 1) |
			((i & 0x8) >> 3))
		table4brev[i] = byte(k)
	}
	table6b5b = make([]inverseEntry, 64)
	for i, row := range table5b6b {
		for _, e := range row {
			table6b5b[int(e)] = inverseEntry{
				valid: true,
				val:   byte(i),
			}
		}
	}
	table4b3b = make([]inverseEntry, 16)
	for i, row := range table3b4b {
		for _, e := range row {
			table4b3b[int(e)] = inverseEntry{
				valid: true,
				val:   byte(i),
			}
		}
	}
}

type inverseEntry struct {
	valid bool
	val   byte
}

var (
	table5b6b = [][]byte{
		/* D.00 .. D.15 */
		{0b100111, 0b011000},
		{0b011101, 0b100010},
		{0b101101, 0b010010},
		{0b110001},
		{0b110101, 0b001010},
		{0b101001},
		{0b011001},
		{0b111000, 0b000111},
		{0b111001, 0b000110},
		{0b100101},
		{0b010101},
		{0b110100},
		{0b001101},
		{0b101100},
		{0b011100},
		{0b010111, 0b101000},
		/* D.16 .. D.31 */
		{0b011011, 0b100100},
		{0b100011},
		{0b010011},
		{0b110010},
		{0b001011},
		{0b101010},
		{0b011010},
		{0b111010, 0b000101},
		{0b110011, 0b001100},
		{0b100110},
		{0b010110},
		{0b110110, 0b001001},
		{0b001110},
		{0b101110, 0b010001},
		{0b011110, 0b100001},
		{0b101011, 0b010100},
	}
	table6b5b []inverseEntry

	table3b4b = [][]byte{
		{0b1011, 0b0100},
		{0b1001},
		{0b0101},
		{0b1100, 0b0011},
		{0b1101, 0b0010},
		{0b1010},
		{0b0110},
		{0b1110, 0b0001, 0b0111, 0b1000}, // special case, yay!
	}
	table4b3b []inverseEntry

	table6brev []byte
	table4brev []byte
)

var (
	_ BitEncoder = &encoder8b10b{}
)
