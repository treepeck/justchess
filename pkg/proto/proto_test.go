package proto

import (
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	for x := range 0xFFFF {
		encoded := EncodeInt(uint16(x))
		decoded := DecodeInt(encoded)
		if decoded != uint16(x) {
			t.Fatalf("expected %d\tgot %d\n", x, decoded)
		}
	}
}

func BenchmarkEncodeInt(b *testing.B) {
	for b.Loop() {
		EncodeInt(0xFFFF)
	}
}

func BenchmarkDecodeInt(b *testing.B) {
	for b.Loop() {
		DecodeInt([]byte{0x7F, 0x7F, 0x3F})
	}
}
