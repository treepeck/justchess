package compression

import (
	"fmt"
	"math/bits"
	"strings"
	"testing"
)

func TestBitWriter(t *testing.T) {
	bw := &bitWriter{remainingBits: intSize}

	expected := 0
	for i := 1; i <= 64; i++ {
		size := bits.Len64(uint64(i))
		expected += size
		bw.write(uint(i), size)
	}

	got := bw.buff.Len()*8 + (intSize - bw.remainingBits)
	if got != expected {
		t.Fatalf("Expected %d bits. Buffer has %d bits\n", expected, got)
	}

	expectedBuff := [41]byte{
		0b11011100, 0b10111011, 0b11000100, 0b11010101, 0b11100110, 0b11110111,
		0b11000010, 0b00110010, 0b10011101, 0b00101011, 0b01101011, 0b11100011,
		0b00111010, 0b11011111, 0b00111011, 0b11101111, 0b11000001, 0b00001100,
		0b01010001, 0b11001001, 0b00101100, 0b11010011, 0b11010001, 0b01001101,
		0b01010101, 0b11011001, 0b01101101, 0b11010111, 0b11100001, 0b10001110,
		0b01011001, 0b11101001, 0b10101110, 0b11011011, 0b11110001, 0b11001111,
		0b01011101, 0b11111001, 0b11101111, 0b11011111, 0b11000000,
	}
	for i, b := range bw.content() {
		if expectedBuff[i] != b {
			t.Fatalf("Expected %b, got %b", expectedBuff[i], b)
		}
	}
}

func TestWriteCompressed(t *testing.T) {
	bw := &bitWriter{remainingBits: intSize}

	for _, n := range []int{-4, -5, -5, -5, -9, -11, -5, -12, -2, -2, -1, -1, -2, -2, -2, -1, -68} {
		bw.writeCompressed(n)
	}

	expected := []byte{
		0x1C, 0x92, 0x49, 0x45, 0x52, 0x57, 0x0C, 0x30,
		0x41, 0x0C, 0x30, 0xC1, 0x9D, 0x00,
	}
	got := bw.content()

	for i, b := range got {
		if expected[i] != b {
			t.Fatalf("Expected %x, got %x", expected, got)
		}
	}
}

func TestBitReader(t *testing.T) {
	cases := []struct {
		input     []byte
		chunkSize int
		expected  string
	}{
		{
			[]byte{
				0b11011100, 0b10111011, 0b11000100, 0b11010101, 0b11100110, 0b11110111,
				0b11000010, 0b00110010, 0b10011101, 0b00101011, 0b01101011, 0b11100011,
				0b00111010, 0b11011111, 0b00111011, 0b11101111, 0b11000001, 0b00001100,
				0b01010001, 0b11001001, 0b00101100, 0b11010011, 0b11010001, 0b01001101,
				0b01010101, 0b11011001, 0b01101101, 0b11010111, 0b11100001, 0b10001110,
				0b01011001, 0b11101001, 0b10101110, 0b11011011, 0b11110001, 0b11001111,
				0b01011101, 0b11111001, 0b11101111, 0b11011111, 0b11000000,
			}, 1,
			"1101110010111011110001001101010111100110111101111100001000110010100111010010101101101011111000110011101011011111001110111110111111000001000011000101000111001001001011001101001111010001010011010101010111011001011011011101011111100001100011100101100111101001101011101101101111110001110011110101110111111001111011111101111111000000",
		},
		{
			[]byte{
				0b11111111, 0b11111111, 0b11111111, 0b11111111, 0b11111111,
				0b11111111, 0b11111111, 0b11111111, 0b11111111, 0b11111111,
			}, 6,
			"111111111111111111111111111111111111111111111111111111111111111111111111111111",
		},
	}

	for i, tc := range cases {
		br := &bitReader{buff: tc.input}

		var got strings.Builder
		for range len(tc.input) * 8 / tc.chunkSize {
			fmt.Fprintf(&got, "%b", br.read(tc.chunkSize))
		}

		if got.String() != tc.expected {
			t.Fatalf("case %d: expected: %v, got: %v", i, tc.expected, got.String())
		}
	}
}

func TestReadCompressed(t *testing.T) {
	br := &bitReader{buff: []byte{
		0x1C, 0x92, 0x49, 0x45, 0x52, 0x57, 0x0C, 0x30,
		0x41, 0x0C, 0x30, 0xC1, 0x9D, 0x00,
	}}

	var got []int
	for range 17 {
		got = append(got, br.readCompressed())
	}

	expected := []int{-4, -5, -5, -5, -9, -11, -5, -12, -2, -2, -1, -1, -2, -2, -2, -1, -68}
	for i, n := range expected {
		if got[i] != n {
			t.Fatalf("expected: %x, got: %x", n, got[i])
		}
	}
}

func BenchmarkBitWriter(b *testing.B) {
	bw := &bitWriter{remainingBits: intSize}
	var i uint
	for b.Loop() {
		bw.write(i, bits.Len64(uint64(i)))
		i++
	}
}

func BenchmarkBitReader(b *testing.B) {
	br := &bitReader{}

	for i := range 255 {
		br.buff = append(br.buff, byte(i))
	}

	for b.Loop() {
		br.read(1)
	}
}
