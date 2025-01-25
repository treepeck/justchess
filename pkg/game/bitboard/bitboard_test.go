package bitboard

import (
	"justchess/pkg/game/enums"
	"justchess/pkg/game/helpers"
	"testing"
)

var dummyBitboard = NewBitboard([2][7]uint64{
	{0x1004EFFD, 0x1000EF00, 0x40040, 0x24, 0x81, 0x8, 0x10},
	{0x80000000, 0x0, 0x0, 0x0, 0x0, 0x80000000, 0x0}},
	enums.White, [4]bool{true, true, true, true},
	-1, 0, 1)

func TestGenLegalMoves(t *testing.T) {
	testcases := []struct {
		name     string
		bitboard *Bitboard
		expected []helpers.Move
	}{
		{"8/8/8/8/4P2q/2N5/PPPP1PPP/R1BQKBNR w KQkq - 0 1",
			dummyBitboard,
			[]helpers.Move{
				// PAWNS
				helpers.NewMove(enums.A3, enums.A2, enums.Quiet),
				helpers.NewMove(enums.A4, enums.A2, enums.DoublePawnPush),
				helpers.NewMove(enums.B3, enums.B2, enums.Quiet),
				helpers.NewMove(enums.B4, enums.B2, enums.DoublePawnPush),
				helpers.NewMove(enums.D3, enums.D2, enums.Quiet),
				helpers.NewMove(enums.D4, enums.D2, enums.DoublePawnPush),
				helpers.NewMove(enums.G3, enums.G2, enums.Quiet),
				helpers.NewMove(enums.G4, enums.G2, enums.DoublePawnPush),
				helpers.NewMove(enums.H3, enums.H2, enums.Quiet),
				helpers.NewMove(enums.E5, enums.E4, enums.Quiet),
				// KNIGHTS
				helpers.NewMove(enums.E2, enums.G1, enums.Quiet),
				helpers.NewMove(enums.F3, enums.G1, enums.Quiet),
				helpers.NewMove(enums.H3, enums.G1, enums.Quiet),
				helpers.NewMove(enums.B1, enums.C3, enums.Quiet),
				helpers.NewMove(enums.E2, enums.C3, enums.Quiet),
				helpers.NewMove(enums.A4, enums.C3, enums.Quiet),
				helpers.NewMove(enums.B5, enums.C3, enums.Quiet),
				helpers.NewMove(enums.D5, enums.C3, enums.Quiet),
				// BISHOPS
				helpers.NewMove(enums.E2, enums.F1, enums.Quiet),
				helpers.NewMove(enums.D3, enums.F1, enums.Quiet),
				helpers.NewMove(enums.C4, enums.F1, enums.Quiet),
				helpers.NewMove(enums.B5, enums.F1, enums.Quiet),
				helpers.NewMove(enums.A6, enums.F1, enums.Quiet),
				// ROOKS
				helpers.NewMove(enums.B1, enums.A1, enums.Quiet),
				// QUEENS
				helpers.NewMove(enums.E2, enums.D1, enums.Quiet),
				helpers.NewMove(enums.F3, enums.D1, enums.Quiet),
				helpers.NewMove(enums.G4, enums.D1, enums.Quiet),
				helpers.NewMove(enums.H5, enums.D1, enums.Quiet),
				// KING
				helpers.NewMove(enums.E2, enums.E1, enums.Quiet),
			},
		},
		{"3q4/8/8/8/8/8/3p1p2/r3K3 w HAha - 0 1",
			NewBitboard([2][7]uint64{
				{0x10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10},
				{0x800000000002801, 0x2800, 0x0, 0x0, 0x1, 0x800000000000000, 0x0},
			},
				enums.White,
				[4]bool{false, false, false, false},
				-1, 0, 0),
			[]helpers.Move{
				helpers.NewMove(enums.E2, enums.E1, enums.Quiet),
				helpers.NewMove(enums.F2, enums.E1, enums.Capture),
			},
		},
	}
	for _, tc := range testcases {
		got := tc.bitboard.GenLegalMoves()

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

///////////////////////////////////////////////////////////////
//                          KING                             //
///////////////////////////////////////////////////////////////

func TestGenKingLegalMoves(t *testing.T) {
	testcases := []struct {
		name     string
		king     int
		allies   uint64
		enemies  uint64
		attacked uint64
		can00    bool
		can000   bool
		expected []helpers.Move
	}{
		{
			"2qr4/5n2/8/3p4/3K4/3P4/8/8",
			enums.D4,
			uint64(1) << enums.D3,
			0xC20000800000000,
			0xFF0E9D7C54840404,
			false,
			false,
			[]helpers.Move{
				helpers.NewMove(enums.E3, enums.D4, enums.Quiet),
			},
		},
		{
			"2q5/8/8/8/8/8/8/R3K2R",
			enums.E1,
			0x81,
			uint64(1) << enums.C8,
			0x4040404040404,
			true,
			true,
			[]helpers.Move{
				helpers.NewMove(enums.G1, enums.E1, enums.KingCastle),
				helpers.NewMove(enums.D1, enums.E1, enums.Quiet),
				helpers.NewMove(enums.D2, enums.E1, enums.Quiet),
				helpers.NewMove(enums.E2, enums.E1, enums.Quiet),
				helpers.NewMove(enums.F1, enums.E1, enums.Quiet),
				helpers.NewMove(enums.F2, enums.E1, enums.Quiet),
			},
		},
		{
			"5q2/8/8/8/8/8/8/R3K2R",
			enums.E1,
			0x81,
			uint64(1) << enums.F8,
			0x20202020202020,
			true,
			true,
			[]helpers.Move{
				helpers.NewMove(enums.B1, enums.E1, enums.QueenCastle),
				helpers.NewMove(enums.D1, enums.E1, enums.Quiet),
				helpers.NewMove(enums.D2, enums.E1, enums.Quiet),
				helpers.NewMove(enums.E2, enums.E1, enums.Quiet),
			},
		},
		{
			"1r6/8/8/8/8/8/1K5r/8",
			enums.B2,
			0x0,
			0x200000000008000,
			0x20202020202FF02,
			false,
			false,
			[]helpers.Move{
				helpers.NewMove(enums.A3, enums.B2, enums.Quiet),
				helpers.NewMove(enums.A1, enums.B2, enums.Quiet),
				helpers.NewMove(enums.C3, enums.B2, enums.Quiet),
				helpers.NewMove(enums.C1, enums.B2, enums.Quiet),
			},
		},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genKingLegalMoves(tc.king, tc.allies, tc.enemies, tc.attacked,
			tc.can00, tc.can000)

		if len(got) != len(tc.expected) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
		for _, emove := range tc.expected {
			isPresent := false
			for _, gmove := range got {
				if gmove == emove {
					isPresent = true
					break
				}
			}
			if !isPresent {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

// func BenchmarkGenKingLegalMoves(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		genKingLegalMoves(enums.E1,
// 			0x81,
// 			uint64(1)<<enums.C8,
// 			0x4040404040404,
// 			true,
// 			true,
// 		)
// 	}
// }

///////////////////////////////////////////////////////////////
//                          PAWN                             //
///////////////////////////////////////////////////////////////

func TestGenWhitePawnPseudoLegalMoves(t *testing.T) {
	testcases := []struct {
		name     string
		pawn     int
		allies   uint64
		enemies  uint64
		expected []helpers.Move
	}{
		{"8/8/8/8/4P3/5p2/4P3/8", enums.E2, uint64(1) << enums.E4, uint64(1) << enums.F3,
			[]helpers.Move{
				helpers.NewMove(enums.E3, enums.E2, enums.Quiet),
				helpers.NewMove(enums.F3, enums.E2, enums.Capture),
			}},
		{"8/4P3/8/8/8/8/8/8", enums.E7, 0, 0,
			[]helpers.Move{
				helpers.NewMove(enums.E8, enums.E7, enums.Promotion),
			}},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genWhitePawnPseudoLegalMoves(tc.pawn, tc.allies, tc.enemies)

		if len(got) != len(tc.expected) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
		for _, gmove := range got {
			isPresent := false
			for _, emove := range tc.expected {
				if gmove == emove {
					isPresent = true
					break
				}
			}
			if !isPresent {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

func TestGenBlackPawnPseudoLegalMoves(t *testing.T) {
	testcases := []struct {
		name     string
		pawn     int
		allies   uint64
		enemies  uint64
		expected []helpers.Move
	}{
		{"8/4p3/5P2/4p3/8/8/8/8", enums.E7, uint64(1) << enums.E5, uint64(1) << enums.F6,
			[]helpers.Move{
				helpers.NewMove(enums.E6, enums.E7, enums.Quiet),
				helpers.NewMove(enums.F6, enums.E7, enums.Capture),
			}},
		{"8/8/8/8/8/8/4p3/8", enums.E2, 0, 0,
			[]helpers.Move{
				helpers.NewMove(enums.E1, enums.E2, enums.Promotion),
			}},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genBlackPawnPseudoLegalMoves(tc.pawn, tc.allies, tc.enemies)

		if len(got) != len(tc.expected) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
		for _, gmove := range got {
			isPresent := false
			for _, emove := range tc.expected {
				if gmove == emove {
					isPresent = true
					break
				}
			}
			if !isPresent {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

// func BenchmarkGenPawnPseudoLegalMoves(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		genWhitePawnPseudoLegalMoves(enums.E2, uint64(1)<<enums.E4, uint64(1)<<enums.F3)
// 	}
// }

///////////////////////////////////////////////////////////////
//                          KNIGHT                           //
///////////////////////////////////////////////////////////////

func TestGenKnightsMovePattern(t *testing.T) {
	testcases := []struct {
		name     string
		knights  uint64
		expected uint64
	}{
		{"knight_e5", 0x24, 0x28440044280000},
		{"knight_a1", 0x0, 0x20400},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genKnightsMovePattern(uint64(1) << tc.knights)

		if got != tc.expected {
			t.Fatalf("expected: %b, got: %b", tc.expected, got)
		}
	}
}

func TestGenKnightPseudoLegalMoves(t *testing.T) {
	testcases := []struct {
		name     string
		knight   int
		allies   uint64
		enemies  uint64
		expected []helpers.Move
	}{
		{"8/8/8/2r3p1/4N3/8/5P2/8", enums.E4, uint64(1) << enums.F2, 0x4400000000,
			[]helpers.Move{
				helpers.NewMove(enums.D6, enums.E4, enums.Quiet),
				helpers.NewMove(enums.G5, enums.E4, enums.Capture),
				helpers.NewMove(enums.D2, enums.E4, enums.Quiet),
				helpers.NewMove(enums.C3, enums.E4, enums.Quiet),
				helpers.NewMove(enums.C5, enums.E4, enums.Capture),
				helpers.NewMove(enums.F6, enums.E4, enums.Quiet),
				helpers.NewMove(enums.G3, enums.E4, enums.Quiet),
			},
		},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genKnightPseudoLegalMoves(tc.knight, tc.allies, tc.enemies)

		if len(got) != len(tc.expected) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
		for _, gmove := range got {
			isPresent := false
			for _, emove := range tc.expected {
				if gmove == emove {
					isPresent = true
					break
				}
			}
			if !isPresent {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

// func BenchmarkGenKnightPseudoLegalmoves(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		genKnightPseudoLegalMoves(enums.E4, uint64(1)<<enums.F2, 0x4400000000)
// 	}
// }

///////////////////////////////////////////////////////////////
//                          BISHOP                           //
///////////////////////////////////////////////////////////////

func TestGenBishopsMovePattern(t *testing.T) {
	testcases := []struct {
		name     string
		bishops  uint64
		allies   uint64
		enemies  uint64
		expected uint64
	}{
		{"8/8/3r4/8/1B6/R1q5/8/8", uint64(1) << enums.B4, uint64(1) << enums.A3, 0x80000040000, 0x80500050000},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genBishopsMovePattern(tc.bishops, tc.allies|tc.enemies)
		if got != tc.expected {
			t.Fatalf("expected: %b, got: %b", tc.expected, got)
		}
	}
}

func TestGenBishopPseudoLegalMoves(t *testing.T) {
	testcases := []struct {
		name     string
		bishop   int
		allies   uint64
		enemies  uint64
		expected []helpers.Move
	}{
		{"8/8/5r2/8/3B4/4q3/1R6/8", enums.D4, uint64(1) << enums.B2, 0x200000100000,
			[]helpers.Move{
				helpers.NewMove(enums.C3, enums.D4, enums.Quiet),
				helpers.NewMove(enums.C5, enums.D4, enums.Quiet),
				helpers.NewMove(enums.B6, enums.D4, enums.Quiet),
				helpers.NewMove(enums.A7, enums.D4, enums.Quiet),
				helpers.NewMove(enums.E3, enums.D4, enums.Capture),
				helpers.NewMove(enums.E5, enums.D4, enums.Quiet),
				helpers.NewMove(enums.F6, enums.D4, enums.Capture),
			}},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genBishopPseudoLegalMoves(tc.bishop, tc.allies, tc.enemies)

		if len(got) != len(tc.expected) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
		for _, gmove := range got {
			isPresent := false
			for _, emove := range tc.expected {
				if gmove == emove {
					isPresent = true
					break
				}
			}
			if !isPresent {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}

///////////////////////////////////////////////////////////////
//                          ROOK                             //
///////////////////////////////////////////////////////////////

func TestGenRooksMovePattern(t *testing.T) {
	testcases := []struct {
		name     string
		rooks    uint64
		allies   uint64
		enemies  uint64
		expected uint64
	}{
		{"6r1/8/8/8/8/4B1R1/8/6p1", uint64(1) << enums.G3, uint64(1) << enums.E3,
			0x4000000000000040, 0x4040404040B04040},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genRooksMovePattern(tc.rooks, tc.allies|tc.enemies)
		if got != tc.expected {
			t.Fatalf("expected: %b, got: %b", tc.expected, got)
		}
	}
}

func TestGenRookPseudoLegalMoves(t *testing.T) {
	testcases := []struct {
		name     string
		rook     int
		allies   uint64
		enemies  uint64
		expected []helpers.Move
	}{
		{"6r1/8/8/8/8/4B1R1/8/6p1", enums.G3, uint64(1) << enums.E3,
			0x4000000000000040, []helpers.Move{
				helpers.NewMove(enums.F3, enums.G3, enums.Quiet),
				helpers.NewMove(enums.H3, enums.G3, enums.Quiet),
				helpers.NewMove(enums.G8, enums.G3, enums.Capture),
				helpers.NewMove(enums.G7, enums.G3, enums.Quiet),
				helpers.NewMove(enums.G6, enums.G3, enums.Quiet),
				helpers.NewMove(enums.G5, enums.G3, enums.Quiet),
				helpers.NewMove(enums.G4, enums.G3, enums.Quiet),
				helpers.NewMove(enums.G2, enums.G3, enums.Quiet),
				helpers.NewMove(enums.G1, enums.G3, enums.Capture),
			}},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s", tc.name)
		got := genRookPseudoLegalMoves(tc.rook, tc.allies, tc.enemies)

		if len(got) != len(tc.expected) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
		for _, gmove := range got {
			isPresent := false
			for _, emove := range tc.expected {
				if gmove == emove {
					isPresent = true
					break
				}
			}
			if !isPresent {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}
		}
	}
}
