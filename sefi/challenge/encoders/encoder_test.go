package encoders

import (
	"bytes"
	"testing"
)

func TestEncoderRoundTrip(t *testing.T) {
	enc := &encoder8b10b{}
	vectors := []string{
		"hello world",
		"\x00\x00\x00\x00\x00\x00",
		"this is lame",
	}
	for _, v := range vectors {
		e := enc.EncodeBytes([]byte(v))
		d, err := enc.DecodeBytes(e)
		if err != nil {
			t.Fatalf("failed on vector %s: %v", v, err)
		}
		if !bytes.Equal([]byte(v), d) {
			t.Fatalf("%v != %v (%v)", v, d, e)
		}
	}
}
