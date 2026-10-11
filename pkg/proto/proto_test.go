package proto

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestEncode(t *testing.T) {
	got, size := Encode(Message{Kind: KindJoin, Id: "testiddddddd", Payload: json.RawMessage("asdasdas")})
	expected := []byte{2, 97, 115, 100, 97, 115, 100, 97, 115, 1, 116, 101, 115, 116, 105, 100, 100, 100, 100, 100, 100, 100, 10}

	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("expected[%d] == %d\tgot[%d] == %d\n", i, expected[i], i, got[i])
		}
	}
	if size != len(expected) {
		t.Fatalf("expected: %d, got: %d\n", size, len(expected))
	}
}

func TestDecode(t *testing.T) {
	encoded := []byte{2, 97, 115, 100, 97, 115, 100, 97, 115, 1, 116, 101, 115, 116, 105, 100, 100, 100, 100, 100, 100, 100, 10}

	parts := PreallocateDecodeBuff()

	got := Decode(parts, encoded, MessageKind(encoded[0]))
	expected := Message{Kind: KindJoin, Id: "testiddddddd", Payload: json.RawMessage("asdasdas")}
	if got.Kind != expected.Kind || got.Id != expected.Id || slices.Compare(got.Payload, expected.Payload) != 0 {
		t.Fatalf("expected: %v, got: %v\n", expected, got)
	}
}

func BenchmarkEncode(b *testing.B) {
	for b.Loop() {
		Encode(Message{Id: "testtesttest", Kind: KindPing, Payload: json.RawMessage("testpayload")})
	}
}

func BenchmarkDecode(b *testing.B) {
	encoded := []byte{116, 101, 115, 116, 116, 101, 115, 116, 116, 101, 115, 116, 32, 0, 32, 116, 101, 115, 116, 112, 97, 121, 108, 111, 97, 100, 10}
	parts := PreallocateDecodeBuff()

	for b.Loop() {
		Decode(parts, encoded, MessageKind(encoded[0]))
	}
}
