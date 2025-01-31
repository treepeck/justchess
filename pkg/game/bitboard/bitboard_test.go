package bitboard

import (
	"testing"

	"justchess/pkg/game/enums"
)

var dummyBitboard = NewBitboard([12]uint64{
	0x1000EF00, 0x0, 0x40040, 0x0, 0x24, 0x0,
	0x81, 0x0, 0x8, 0x80000000, 0x10, 0x0},
	enums.White, [4]bool{true, true, true, true},
	-1, 0, 1)

func TestGenLegalMoves(t *testing.T) {
	testcases := []struct {
		name     string
		bitboard *Bitboard
		expected []Move
	}{
		{"8/8/8/8/4P2q/2N5/PPPP1PPP/R1BQKBNR w KQkq - 0 1",
			dummyBitboard,
			[]Move{
				// PAWNS
				NewMove(enums.A3, enums.A2, enums.Quiet),
				NewMove(enums.A4, enums.A2, enums.DoublePawnPush),
				NewMove(enums.B3, enums.B2, enums.Quiet),
				NewMove(enums.B4, enums.B2, enums.DoublePawnPush),
				NewMove(enums.D3, enums.D2, enums.Quiet),
				NewMove(enums.D4, enums.D2, enums.DoublePawnPush),
				NewMove(enums.G3, enums.G2, enums.Quiet),
				NewMove(enums.G4, enums.G2, enums.DoublePawnPush),
				NewMove(enums.H3, enums.H2, enums.Quiet),
				NewMove(enums.E5, enums.E4, enums.Quiet),
				// KNIGHTS
				NewMove(enums.E2, enums.G1, enums.Quiet),
				NewMove(enums.F3, enums.G1, enums.Quiet),
				NewMove(enums.H3, enums.G1, enums.Quiet),
				NewMove(enums.B1, enums.C3, enums.Quiet),
				NewMove(enums.E2, enums.C3, enums.Quiet),
				NewMove(enums.A4, enums.C3, enums.Quiet),
				NewMove(enums.B5, enums.C3, enums.Quiet),
				NewMove(enums.D5, enums.C3, enums.Quiet),
				// BISHOPS
				NewMove(enums.E2, enums.F1, enums.Quiet),
				NewMove(enums.D3, enums.F1, enums.Quiet),
				NewMove(enums.C4, enums.F1, enums.Quiet),
				NewMove(enums.B5, enums.F1, enums.Quiet),
				NewMove(enums.A6, enums.F1, enums.Quiet),
				// ROOKS
				NewMove(enums.B1, enums.A1, enums.Quiet),
				// QUEENS
				NewMove(enums.E2, enums.D1, enums.Quiet),
				NewMove(enums.F3, enums.D1, enums.Quiet),
				NewMove(enums.G4, enums.D1, enums.Quiet),
				NewMove(enums.H5, enums.D1, enums.Quiet),
				// KING
				NewMove(enums.E2, enums.E1, enums.Quiet),
			},
		},
		{"3q4/8/8/8/8/8/3p1p2/r3K3 w HAha - 0 1",
			NewBitboard([12]uint64{
				0x0, 0x2800, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x800000000000000,
				0x10, 0x0,
			},
				enums.White,
				[4]bool{false, false, false, false},
				-1, 0, 0),
			[]Move{
				NewMove(enums.E2, enums.E1, enums.Quiet),
				NewMove(enums.F2, enums.E1, enums.Capture),
			},
		},
	}
	for _, tc := range testcases {
		tc.bitboard.GenLegalMoves()
		got := tc.bitboard.LegalMoves

		if len(tc.expected) != len(got) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
		for i, move := range tc.expected {
			if got[i] != move {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

func BenchmarkGenLegalMoves(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dummyBitboard.GenLegalMoves()
	}
}
