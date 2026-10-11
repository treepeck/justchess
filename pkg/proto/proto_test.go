package proto

import (
	"slices"
	"testing"
)

func TestEncode(t *testing.T) {
	got := Encode(Message{Kind: KindJoin, Id: "testiddddddd", Payload: []byte("asdasdas")})
	expected := []byte{2, 116, 101, 115, 116, 105, 100, 100, 100, 100, 100, 100, 100, 97, 115, 100, 97, 115, 100, 97, 115, MessageSeparator}
	if slices.Compare(got[:], expected[:]) != 0 {
		t.Fatalf("expected %v\tgot %v\n", expected, got)
	}
}

func TestDecode(t *testing.T) {
	encoded := []byte{2, 116, 101, 115, 116, 105, 100, 100, 100, 100, 100, 100, 100, 97, 115, 100, 97, 115, 100, 97, 115, MessageSeparator}

	got := Decode(encoded)
	expected := Message{Kind: KindJoin, Id: "testiddddddd", Payload: []byte("asdasdas")}
	if got.Kind != expected.Kind || got.Id != expected.Id || slices.Compare(got.Payload, expected.Payload) != 0 {
		t.Fatalf("expected: %v, got: %v\n", expected, got)
	}
}

func BenchmarkEncode(b *testing.B) {
	for b.Loop() {
		Encode(Message{Id: "testtesttest", Kind: KindPing, Payload: []byte("testpayload")})
	}
}

func BenchmarkDecode(b *testing.B) {
	encoded := []byte{2, 116, 101, 115, 116, 105, 100, 100, 100, 100, 100, 100, 100, 97, 115, 100, 97, 115, 100, 97, 115, MessageSeparator}

	for b.Loop() {
		Decode(encoded)
	}
}
