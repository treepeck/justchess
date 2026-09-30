package proto

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestEncode(t *testing.T) {
	e := NewEncoder()

	// Heartbeat message.
	got := e.Encode(Message{Kind: KindPing})
	if got != nil {
		t.Fatalf("expected: %v, got: %v\n", nil, got)
	}

	// General message.
	got = e.Encode(Message{Kind: KindJoin, Id: "testiddddddd", Payload: json.RawMessage("asdasdas")})
	expected := []byte{2, 97, 115, 100, 97, 115, 100, 97, 115, 1, 116, 101, 115, 116, 105, 100, 100, 100, 100, 100, 100, 100, 10}
	if slices.Compare(got, expected) != 0 {
		t.Fatalf("expected: %v, got: %v\n", expected, got)
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
	e := NewEncoder()

	for b.Loop() {
		e.Encode(Message{Id: "testtesttest", Kind: KindPing, Payload: json.RawMessage("testpayload")})
	}
}

func BenchmarkDecode(b *testing.B) {
	encoded := []byte{116, 101, 115, 116, 116, 101, 115, 116, 116, 101, 115, 116, 32, 0, 32, 116, 101, 115, 116, 112, 97, 121, 108, 111, 97, 100, 10}
	parts := PreallocateDecodeBuff()

	for b.Loop() {
		Decode(parts, encoded, MessageKind(encoded[0]))
	}
}
