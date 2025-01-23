package fen

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"testing"
)

var dummyBitboard = bitboard.NewBitboard([2][7]uint64{
	{0xB25CC69, 0xA21C400, 0x40000, 0x1000800, 0x21, 0x8, 0x40},
	{0x91CB242910000000, 0xC3040810000000, 0x8002000000000, 0x0100000000, 0x8100000000000000, 0x200000000000, 0x1000000000000000},
}, enums.White, [4]bool{false, false, true, true},
	enums.B3, 0, 13)

var dummyFEN = []string{
	"r3k2r/pp1n2pp/2p2q2/b2p1n2/BP1Pp3/P1N2P2/2PB2PP/R2Q1RK1 w kq b3 0 13",
	"8/8/8/8/4P2q/2N5/PPPP1PPP/R1BQKBNR w KQkq - 0 1",
}

func TestBitboard2FEN(t *testing.T) {
	testcases := []struct {
		bitboard *bitboard.Bitboard
		expected string
	}{
		{
			dummyBitboard,
			"r3k2r/pp1n2pp/2p2q2/b2p1n2/BP1Pp3/P1N2P2/2PB2PP/R2Q1RK1 w kq b3 0 13",
		},
	}
	for _, tc := range testcases {
		got := Bitboard2FEN(tc.bitboard)

		if tc.expected != got {
			t.Fatalf("expected: %s, got: %s", tc.expected, got)
		}
	}
}

func BenchmarkBitboard2FEN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Bitboard2FEN(dummyBitboard)
	}
}

func TestFEN2Bitboard(t *testing.T) {
	testcases := []struct {
		FEN                    string
		expectedPieces         [2][7]uint64
		expectedActiveColor    enums.Color
		expectedCastlingRights [4]bool
		expectedEpTarget       int
		expectedHalfmoveCLK    int
		expectedFullmoveCLK    int
	}{
		{
			dummyFEN[0],
			[2][7]uint64{
				{0xB25CC69, 0xA21C400, 0x40000, 0x1000800, 0x21, 0x8, 0x40},
				{0x91CB242910000000, 0xC3040810000000, 0x8002000000000, 0x0100000000, 0x8100000000000000, 0x200000000000, 0x1000000000000000}},
			enums.White, [4]bool{false, false, true, true},
			enums.B3, 0, 13,
		},
		{
			dummyFEN[1],
			[2][7]uint64{
				{0x1004EFFD, 0x1000EF00, 0x40040, 0x24, 0x81, 0x8, 0x10},
				{0x80000000, 0x0, 0x0, 0x0, 0x0, 0x80000000, 0x0}},
			enums.White, [4]bool{true, true, true, true},
			-1, 0, 1,
		},
	}
	for _, tc := range testcases {
		got := FEN2Bitboard(tc.FEN)
		for i, pieces := range got.Pieces {
			for j, piece := range pieces {
				if piece != tc.expectedPieces[i][j] {
					t.Fatalf("expected: %v, got: %v", tc.expectedPieces, got.Pieces)
				}
			}
		}
		if tc.expectedActiveColor != got.ActiveColor {
			t.Fatalf("expected: %d, got: %d", tc.expectedActiveColor, got.ActiveColor)
		}
		for i, bb := range got.CastlingRights {
			if bb != tc.expectedCastlingRights[i] {
				t.Fatalf("expected: %v, got: %v", tc.expectedCastlingRights, got.CastlingRights)
			}
		}
		if tc.expectedEpTarget != got.EPTarget {
			t.Fatalf("expected: %d, got: %d", tc.expectedEpTarget, got.EPTarget)
		}
		if tc.expectedHalfmoveCLK != got.HalfmoveCnt {
			t.Fatalf("expected: %d, got: %d", tc.expectedHalfmoveCLK, got.HalfmoveCnt)
		}
		if tc.expectedFullmoveCLK != got.FullmoveCnt {
			t.Fatalf("expected: %d, got: %d", tc.expectedFullmoveCLK, got.FullmoveCnt)
		}
	}
}

func BenchmarkFEN2Bitboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FEN2Bitboard(dummyFEN[0])
	}
}
