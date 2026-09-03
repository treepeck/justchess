package compression

import (
	"testing"
)

func TestHuffmanEncoding(t *testing.T) {
	expected := []byte{
		0b11101110, 0b00001010, 0b01100000,
	}

	got := HuffmanEncoding([]byte{9, 9, 22, 17})

	for i, b := range expected {
		if got[i] != b {
			t.Fatalf("expected: %v, got: %v", expected, got)
		}
	}
}

func TestHuffmanDecoding(t *testing.T) {
	indices := []int{9, 9, 22, 17}

	w := bitWriter{remainingBits: intSize}
	for _, i := range indices {
		w.write(huffmanCodes[i].code, huffmanCodes[i].size)
	}

	expected := []Move{
		{"e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR"},
		{"e5", "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR"},
		{"Bc4", "rnbqkbnr/pppp1ppp/8/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR"},
		{"Nf6", "rnbqkb1r/pppp1ppp/5n2/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR"},
	}

	got := HuffmanDecoding(w.content(), len(indices))

	for i, m := range expected {
		if got[i] != m {
			t.Fatalf("expected: %v, got: %v", expected, got)
		}
	}
}

func BenchmarkMakeTrie(b *testing.B) {
	for b.Loop() {
		makeTrie()
	}
}

func BenchmarkCompressTimeDiffs(b *testing.B) {
	for b.Loop() {
		CompressTimeDiffs([]int{-20, -30, -40, 2, 4, 10, -200, 10})
	}
}

func BenchmarkDecompressTimeDiffs(b *testing.B) {
	for b.Loop() {
		DecompressTimeDiffs([]byte{0b00101001, 0b01001010, 0b00000111,
			0b11000001, 0b10111100, 0b10111000, 0b10001000, 0b11000001}, 6)
	}
}
