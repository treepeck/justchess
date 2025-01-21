package fen

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"testing"
)

func TestBitboard2FEN(t *testing.T) {
	testcases := []struct {
		bitboard *bitboard.Bitboard
		expected string
	}{
		{
			bitboard.NewBitboard([14]uint64{
				0xB25CC69,
				0x91CB242910000000,
				0xA21C400,
				0xC3040810000000,
				0x40000,
				0x8002000000000,
				0x1000800,
				0x0100000000,
				0x21,
				0x8100000000000000,
				0x8,
				0x200000000000,
				0x40,
				0x1000000000000000}, enums.White, [4]bool{false, false, true, true},
				enums.B3, 0, 13),
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

func TestFEN2Bitboard(t *testing.T) {
	testcases := []struct {
		FEN                    string
		expectedPieces         [14]uint64
		expectedActiveColor    enums.Color
		expectedCastlingRights [4]bool
		expectedEpTarget       int
		expectedHalfmoveCLK    int
		expectedFullmoveCLK    int
	}{
		{
			"r3k2r/pp1n2pp/2p2q2/b2p1n2/BP1Pp3/P1N2P2/2PB2PP/R2Q1RK1 w kq b3 0 13",
			[14]uint64{
				0xB25CC69,
				0x91CB242910000000,
				0xA21C400,
				0xC3040810000000,
				0x40000,
				0x8002000000000,
				0x1000800,
				0x0100000000,
				0x21,
				0x8100000000000000,
				0x8,
				0x200000000000,
				0x40,
				0x1000000000000000},
			enums.White, [4]bool{false, false, true, true},
			enums.B3, 0, 13,
		},
	}
	for _, tc := range testcases {
		got := FEN2Bitboard(tc.FEN)
		for i, bb := range got.Pieces {
			if bb != tc.expectedPieces[i] {
				t.Fatalf("expected: %v, got: %v", tc.expectedPieces, got.Pieces)
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
		if tc.expectedEpTarget != got.EpTarget {
			t.Fatalf("expected: %d, got: %d", tc.expectedEpTarget, got.EpTarget)
		}
		if tc.expectedHalfmoveCLK != got.HalfmoveClk {
			t.Fatalf("expected: %d, got: %d", tc.expectedHalfmoveCLK, got.HalfmoveClk)
		}
		if tc.expectedFullmoveCLK != got.FullmoveClk {
			t.Fatalf("expected: %d, got: %d", tc.expectedFullmoveCLK, got.FullmoveClk)
		}
	}
}
