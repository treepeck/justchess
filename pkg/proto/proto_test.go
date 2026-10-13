package proto

import (
	"slices"
	"testing"
)

func TestEncodeDecodeInt(t *testing.T) {
	for x := range 0xFFFF {
		encoded := EncodeInt(uint16(x))
		decoded := DecodeInt(encoded)
		if decoded != uint16(x) {
			t.Fatalf("expected %d\tgot %d\n", x, decoded)
		}
	}
}

func TestEncode(t *testing.T) {
	expected := make([]byte, 0, 10)
	expected = append(expected, KindPing)
	expected = append(expected, EncodeInt(120)...)
	expected = append(expected, MessageSeparator)

	got := Encode(Message{
		Kind:    KindPing,
		Payload: EncodeInt(120),
	})
	if slices.Compare(expected, got) != 0 {
		t.Fatalf("expected %b\tgot %b\n", expected, got)
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

func BenchmarkEncode(b *testing.B) {
	for b.Loop() {
		Encode(Message{
			Kind:    KindPing,
			Id:      "123456789111",
			Payload: EncodeInt(120),
		})
	}
}
