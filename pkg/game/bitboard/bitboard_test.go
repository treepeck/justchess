package bitboard

import (
	"justchess/pkg/game/enums"
	"testing"
)

func TestMakeMove(t *testing.T) {
	testcases := []struct {
		name     string
		before   [12]uint64
		move     Move
		expected [12]uint64
	}{
		{
			"8/4p3/1PR5/8/4R3/8/4p3/8 w - - 0 1 | Quiet Re3",
			[12]uint64{
				0x20000000000, 0x10000000001000, 0x0, 0x0, 0x0, 0x0,
				0x40010000000, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			NewMove(enums.E3, enums.E4, enums.Quiet),
			[12]uint64{
				0x20000000000, 0x10000000001000, 0x0, 0x0, 0x0, 0x0,
				0x40000100000, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
		},
		{
			"8/4p3/1PR5/8/4R3/8/4p3/8 w - - 0 1 | Capture Rxe7",
			[12]uint64{
				0x20000000000, 0x10000000001000, 0x0, 0x0, 0x0, 0x0,
				0x40010000000, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			NewMove(enums.E7, enums.E4, enums.Capture),
			[12]uint64{
				0x20000000000, 0x1000, 0x0, 0x0, 0x0, 0x0,
				0x10040000000000, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
		},
		{
			"4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1 | White O-O",
			[12]uint64{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x81, 0x0, 0x0, 0x0, 0x10, 0x1000000000000000,
			},
			NewMove(enums.G1, enums.E1, enums.KingCastle),
			[12]uint64{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x21, 0x0, 0x0, 0x0, 0x40, 0x1000000000000000,
			},
		},
		{
			"r3k3/8/8/8/8/8/8/R4RK1 b q - 0 1 | Black O-O-O",
			[12]uint64{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x21, 0x100000000000000, 0x0, 0x0, 0x40, 0x1000000000000000,
			},
			NewMove(enums.C8, enums.E8, enums.QueenCastle),
			[12]uint64{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x21, 0x800000000000000, 0x0, 0x0, 0x40, 0x400000000000000,
			},
		},
		{
			"4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1 | EnPassant exd6",
			[12]uint64{
				0x1000000000, 0x800000000, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x10, 0x1000000000000000,
			},
			NewMove(enums.D6, enums.E5, enums.EPCapture),
			[12]uint64{
				0x80000000000, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x10, 0x1000000000000000,
			},
		},
		{
			"4k3/8/8/4P3/8/8/2p5/6K1 b - - 0 1 | Promo c1=Q",
			[12]uint64{
				0x1000000000, 0x400, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x10, 0x1000000000000000,
			},
			NewMove(enums.C1, enums.C2, enums.QueenPromo),
			[12]uint64{
				0x1000000000, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x4, 0x10, 0x1000000000000000,
			},
		},
		{
			"4k3/8/8/4P3/8/8/2p5/1Q4K1 b - - 0 1 | PromoCapture cxb1=N",
			[12]uint64{
				0x1000000000, 0x400, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x2, 0x0, 0x10, 0x1000000000000000,
			},
			NewMove(enums.B1, enums.C2, enums.KnightPromoCapture),
			[12]uint64{
				0x1000000000, 0x0, 0x0, 0x2, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x10, 0x1000000000000000,
			},
		},
	}

	bb := NewBitboard([12]uint64{}, enums.White, [4]bool{false, false, false, false},
		enums.NoSquare, 0, 0)
	for _, tc := range testcases {
		t.Logf("passing test: %s\n", tc.name)
		bb.Pieces = tc.before
		bb.MakeMove(tc.move)

		for pt, board := range bb.Pieces {
			if tc.expected[pt] != board {
				t.Fatalf("expected: %v, got: %v\n", tc.expected, bb.Pieces)
			}
		}
	}
}

func BenchmarkMakeMove(b *testing.B) {
	bb := NewBitboard([12]uint64{0x1000000000, 0x0, 0x0, 0x2, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x10, 0x1000000000000000}, enums.White, [4]bool{false, false, false, false},
		enums.NoSquare, 0, 0)
	for i := 0; i < b.N; i++ {
		bb.MakeMove(NewMove(enums.B1, enums.C2, enums.KnightPromoCapture))
	}
}

func TestGenLegalMoves(t *testing.T) {
	testcases := []struct {
		fen      string
		bitboard *Bitboard
		expected []Move
	}{
		{"8/8/8/8/4P2q/2N5/PPPP1PPP/R1BQKBNR w KQkq - 0 1",
			NewBitboard([12]uint64{
				0x1000EF00, 0x0, 0x40040, 0x0, 0x24, 0x0,
				0x81, 0x0, 0x8, 0x80000000, 0x10, 0x0,
			}, enums.White, [4]bool{true, true, true, true}, enums.NoSquare, 0, 1),
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
				NewMove(enums.C8, enums.E8, enums.QueenCastle),
			},
		},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s\n", tc.fen)
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
	bb := NewBitboard([12]uint64{
		0x1000EF00, 0x0, 0x40040, 0x0, 0x24, 0x0,
		0x81, 0x0, 0x8, 0x80000000, 0x10, 0x0,
	}, enums.White, [4]bool{true, true, true, true}, enums.NoSquare, 0, 1)
	for i := 0; i < b.N; i++ {
		bb.GenLegalMoves()
	}
}

func TestKingGenLegalMoves(t *testing.T) {
	testcases := []struct {
		fen      string
		king     uint64
		allies   uint64
		enemies  uint64
		attacked uint64
		color    enums.Color
		canOO    bool
		canOOO   bool
		expected []Move
	}{
		{"r2qk2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			0x10, 0x81, 0x9900000000000000, 0xFFBDABC989898989,
			enums.White, true, true, []Move{
				NewMove(enums.F1, enums.E1, enums.Quiet),
				NewMove(enums.E2, enums.E1, enums.Quiet),
				NewMove(enums.F2, enums.E1, enums.Quiet),
				NewMove(enums.G1, enums.E1, enums.KingCastle),
			},
		},
		{
			"r3k2r/8/8/8/8/8/8/8 b kq - 0 1",
			0x1000000000000000, 0x8100000000000000, 0x0, 0x0,
			enums.Black, true, true, []Move{
				NewMove(enums.D7, enums.E8, enums.Quiet),
				NewMove(enums.E7, enums.E8, enums.Quiet),
				NewMove(enums.F7, enums.E8, enums.Quiet),
				NewMove(enums.D8, enums.E8, enums.Quiet),
				NewMove(enums.F8, enums.E8, enums.Quiet),
				NewMove(enums.G8, enums.E8, enums.KingCastle),
				NewMove(enums.C8, enums.E8, enums.QueenCastle),
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("passing test: %s\n", tc.fen)

		got := genKingLegalMoves(tc.king, tc.allies, tc.enemies, tc.attacked,
			tc.canOO, tc.canOOO, tc.color)

		for i, move := range tc.expected {
			if move != got[i] {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

func BenchmarkGenKingLegalMoves(b *testing.B) {
	for i := 0; i < b.N; i++ {
		genKingLegalMoves(0x10, 0x81, 0x9900000000000000, 0xFFBDABC989898989,
			true, true, enums.White)
	}
}

func TestPawnsPseudoLegalMoves(t *testing.T) {
	testcases := []struct {
		fen      string
		pawns    uint64
		allies   uint64
		enemies  uint64
		expected []Move
		epTarget int
		color    enums.Color
	}{
		{
			"8/8/8/8/8/8/PPPP4/8 w - - 0 1",
			0xF00, 0x0, 0x0,
			[]Move{
				NewMove(enums.A3, enums.A2, enums.Quiet),
				NewMove(enums.A4, enums.A2, enums.DoublePawnPush),
				NewMove(enums.B3, enums.B2, enums.Quiet),
				NewMove(enums.B4, enums.B2, enums.DoublePawnPush),
				NewMove(enums.C3, enums.C2, enums.Quiet),
				NewMove(enums.C4, enums.C2, enums.DoublePawnPush),
				NewMove(enums.D3, enums.D2, enums.Quiet),
				NewMove(enums.D4, enums.D2, enums.DoublePawnPush),
			}, enums.NoSquare, enums.White,
		},
		{
			"8/8/8/3ppp2/4P3/8/8/8 w - - 0 1",
			0x10000000, 0x0, 0x3800000000,
			[]Move{
				NewMove(enums.D5, enums.E4, enums.Capture),
				NewMove(enums.F5, enums.E4, enums.Capture),
			}, enums.NoSquare, enums.White,
		},
		{
			"8/8/8/8/3pP3/8/8/8 b - e3 0 1",
			0x8000000, 0x0, 0x10000000,
			[]Move{
				NewMove(enums.D3, enums.D4, enums.Quiet),
				NewMove(enums.E3, enums.D4, enums.EPCapture),
			}, enums.E3, enums.Black,
		},
		{
			"8/8/8/8/8/8/3p4/2B5 b - - 0 1",
			0x800, 0x0, 0x4,
			[]Move{
				NewMove(enums.D1, enums.D2, enums.QueenPromo),
				NewMove(enums.C1, enums.D2, enums.QueenPromoCapture),
			}, enums.NoSquare, enums.Black,
		},
	}

	for _, tc := range testcases {
		t.Logf("passing test: %s\n", tc.fen)
		got := genPawnsPseudoLegalMoves(tc.pawns, tc.allies, tc.enemies, tc.color, tc.epTarget)

		for i, move := range tc.expected {
			if move != got[i] {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

func BenchmarkGenPawnsPseudoLegalMoves(b *testing.B) {
	for i := 0; i < b.N; i++ {
		genPawnsPseudoLegalMoves(0xF00|0x10000000|0x8000000|0x800, 0x0,
			0x3800000000|0x10000000|0x4, enums.White, enums.NoSquare)
	}
}

func TestGenPseudoLegalMoves(t *testing.T) {
	testcases := []struct {
		fen       string
		pieceType enums.PieceType
		bb        uint64
		allies    uint64
		enemies   uint64
		expected  []Move
	}{
		{
			"knights: 8/3p4/8/2N5/6N1/8/5p2/8 w - - 0 1", enums.WhiteKnight,
			0x440000000, 0x440000000, 0x8000000002000,
			[]Move{
				NewMove(enums.F2, enums.G4, enums.Capture),
				NewMove(enums.H2, enums.G4, enums.Quiet),
				NewMove(enums.E3, enums.G4, enums.Quiet),
				NewMove(enums.E5, enums.G4, enums.Quiet),
				NewMove(enums.F6, enums.G4, enums.Quiet),
				NewMove(enums.H6, enums.G4, enums.Quiet),
				NewMove(enums.B3, enums.C5, enums.Quiet),
				NewMove(enums.D3, enums.C5, enums.Quiet),
				NewMove(enums.A4, enums.C5, enums.Quiet),
				NewMove(enums.E4, enums.C5, enums.Quiet),
				NewMove(enums.A6, enums.C5, enums.Quiet),
				NewMove(enums.E6, enums.C5, enums.Quiet),
				NewMove(enums.B7, enums.C5, enums.Quiet),
				NewMove(enums.D7, enums.C5, enums.Capture),
			},
		},
		{
			"bishops: 8/4p3/1P6/2B5/6B1/8/4p3/8 w - - 0 1", enums.WhiteBishop,
			0x440000000, 0x20000000000, 0x10000000001000,
			[]Move{
				NewMove(enums.E2, enums.G4, enums.Capture),
				NewMove(enums.F3, enums.G4, enums.Quiet),
				NewMove(enums.H3, enums.G4, enums.Quiet),
				NewMove(enums.F5, enums.G4, enums.Quiet),
				NewMove(enums.H5, enums.G4, enums.Quiet),
				NewMove(enums.E6, enums.G4, enums.Quiet),
				NewMove(enums.D7, enums.G4, enums.Quiet),
				NewMove(enums.C8, enums.G4, enums.Quiet),
				NewMove(enums.G1, enums.C5, enums.Quiet),
				NewMove(enums.F2, enums.C5, enums.Quiet),
				NewMove(enums.A3, enums.C5, enums.Quiet),
				NewMove(enums.E3, enums.C5, enums.Quiet),
				NewMove(enums.B4, enums.C5, enums.Quiet),
				NewMove(enums.D4, enums.C5, enums.Quiet),
				NewMove(enums.D6, enums.C5, enums.Quiet),
				NewMove(enums.E7, enums.C5, enums.Capture),
			},
		},
		{
			"rooks: 8/4p3/1P6/2B5/6B1/8/4p3/8 w - - 0 1", enums.WhiteRook,
			0x40010000000, 0x20000000000, 0x10000000001000,
			[]Move{
				NewMove(enums.E2, enums.E4, enums.Capture),
				NewMove(enums.E3, enums.E4, enums.Quiet),
				NewMove(enums.A4, enums.E4, enums.Quiet),
				NewMove(enums.B4, enums.E4, enums.Quiet),
				NewMove(enums.C4, enums.E4, enums.Quiet),
				NewMove(enums.D4, enums.E4, enums.Quiet),
				NewMove(enums.F4, enums.E4, enums.Quiet),
				NewMove(enums.G4, enums.E4, enums.Quiet),
				NewMove(enums.H4, enums.E4, enums.Quiet),
				NewMove(enums.E5, enums.E4, enums.Quiet),
				NewMove(enums.E6, enums.E4, enums.Quiet),
				NewMove(enums.E7, enums.E4, enums.Capture),
				NewMove(enums.C1, enums.C6, enums.Quiet),
				NewMove(enums.C2, enums.C6, enums.Quiet),
				NewMove(enums.C3, enums.C6, enums.Quiet),
				NewMove(enums.C4, enums.C6, enums.Quiet),
				NewMove(enums.C5, enums.C6, enums.Quiet),
				NewMove(enums.D6, enums.C6, enums.Quiet),
				NewMove(enums.E6, enums.C6, enums.Quiet),
				NewMove(enums.F6, enums.C6, enums.Quiet),
				NewMove(enums.G6, enums.C6, enums.Quiet),
				NewMove(enums.H6, enums.C6, enums.Quiet),
				NewMove(enums.C7, enums.C6, enums.Quiet),
				NewMove(enums.C8, enums.C6, enums.Quiet),
			},
		},
	}

	for _, tc := range testcases {
		t.Logf("passing test: %s\n", tc.fen)

		got := genPseudoLegalMoves(tc.pieceType, tc.bb, tc.allies, tc.enemies)
		for i, move := range tc.expected {
			if move != got[i] {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

func BenchmarkGenPseudoLegalMoves(b *testing.B) {
	for i := 0; i < b.N; i++ {
		//genPseudoLegalMoves(enums.WhiteKnight, 0x440000000, 0x440000000, 0x8000000002000)
		genPseudoLegalMoves(enums.BlackQueen, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0x0)
	}
}
