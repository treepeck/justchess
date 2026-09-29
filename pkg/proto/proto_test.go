package proto

import (
	"encoding/json"
	"testing"
)

func TestDecoder(t *testing.T) {
	raw := []byte{116, 101, 115, 116, 116, 101, 115, 116, 116, 101, 115, 116, 32, 0, 32, 116, 101, 115, 116, 112, 97, 121, 108, 111, 97, 100}
	parts := make([][]byte, 3)
	parts[0] = make([]byte, 12)  // Enough to store the ID.
	parts[1] = make([]byte, 1)   // Enough to store the MessageKind.
	parts[2] = make([]byte, 100) // Enough to store every possible Payload.
	m, err := Decode(parts, raw)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("message: %v\n", m)
}

func BenchmarkEncode(b *testing.B) {
	e := NewEncoder()

	for b.Loop() {
		e.Encode(Message{Id: "testtesttest", Kind: KindPing, Payload: json.RawMessage("testpayload")})
	}
}

func BenchmarkDecode(b *testing.B) {
	raw := []byte{116, 101, 115, 116, 116, 101, 115, 116, 116, 101, 115, 116, 32, 0, 32, 116, 101, 115, 116, 112, 97, 121, 108, 111, 97, 100}
	parts := make([][]byte, 3)
	parts[0] = make([]byte, 12)  // Enough to store the ID.
	parts[1] = make([]byte, 1)   // Enough to store the MessageKind.
	parts[2] = make([]byte, 100) // Enough to store every possible Payload.

	for b.Loop() {
		Decode(parts, raw)
	}
}
