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
		{"2q1k3/8/8/8/8/8/8/R3K2R w KQ - 0 1",
			NewBitboard([12]uint64{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80, 0x0, 0x0, 0x400000000000000, 0x10, 0x1000000000000000,
			}, enums.White, [4]bool{true, false, true, false}, -1, 0, 1),
			[]Move{
				NewMove(enums.F1, enums.H1, enums.Quiet),
				NewMove(enums.G1, enums.H1, enums.Quiet),
				NewMove(enums.H2, enums.H1, enums.Quiet),
				NewMove(enums.H3, enums.H1, enums.Quiet),
				NewMove(enums.H4, enums.H1, enums.Quiet),
				NewMove(enums.H5, enums.H1, enums.Quiet),
				NewMove(enums.H6, enums.H1, enums.Quiet),
				NewMove(enums.H7, enums.H1, enums.Quiet),
				NewMove(enums.H8, enums.H1, enums.Quiet),
				NewMove(enums.D1, enums.E1, enums.Quiet),
				NewMove(enums.F1, enums.E1, enums.Quiet),
				NewMove(enums.D2, enums.E1, enums.Quiet),
				NewMove(enums.E2, enums.E1, enums.Quiet),
				NewMove(enums.F2, enums.E1, enums.Quiet),
				NewMove(enums.G1, enums.E1, enums.KingCastle),
			},
		},
		{"r3k2r/8/8/8/8/8/8/4K1R1 w kq - 0 1",
			NewBitboard([12]uint64{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x40, 0x8100000000000000, 0x0, 0x0, 0x10, 0x1000000000000000,
			}, enums.Black, [4]bool{false, true, false, true}, -1, 0, 1),
			[]Move{
				NewMove(enums.A1, enums.A8, enums.Quiet),
				NewMove(enums.A2, enums.A8, enums.Quiet),
				NewMove(enums.A3, enums.A8, enums.Quiet),
				NewMove(enums.A4, enums.A8, enums.Quiet),
				NewMove(enums.A5, enums.A8, enums.Quiet),
				NewMove(enums.A6, enums.A8, enums.Quiet),
				NewMove(enums.A7, enums.A8, enums.Quiet),
				NewMove(enums.B8, enums.A8, enums.Quiet),
				NewMove(enums.C8, enums.A8, enums.Quiet),
				NewMove(enums.D8, enums.A8, enums.Quiet),
				NewMove(enums.H1, enums.H8, enums.Quiet),
				NewMove(enums.H2, enums.H8, enums.Quiet),
				NewMove(enums.H3, enums.H8, enums.Quiet),
				NewMove(enums.H4, enums.H8, enums.Quiet),
				NewMove(enums.H5, enums.H8, enums.Quiet),
				NewMove(enums.H6, enums.H8, enums.Quiet),
				NewMove(enums.H7, enums.H8, enums.Quiet),
				NewMove(enums.F8, enums.H8, enums.Quiet),
				NewMove(enums.G8, enums.H8, enums.Quiet),
				NewMove(enums.D7, enums.E8, enums.Quiet),
				NewMove(enums.E7, enums.E8, enums.Quiet),
				NewMove(enums.F7, enums.E8, enums.Quiet),
				NewMove(enums.D8, enums.E8, enums.Quiet),
				NewMove(enums.F8, enums.E8, enums.Quiet),
				NewMove(enums.B8, enums.E8, enums.QueenCastle),
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
