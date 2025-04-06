package chess

import (
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
	"justchess/pkg/chess/fen"
	"testing"

	"github.com/google/uuid"
)

func TestProcessMove(t *testing.T) {
	testcases := []struct {
		san         string
		beforeFEN   string
		move        bitboard.Move
		expectedFEN string
	}{
		{
			"e4",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			bitboard.NewMove(enums.E4, enums.E2, enums.DoublePawnPush),
			"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		},
		{
			"exd5",
			"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2",
			bitboard.NewMove(enums.D5, enums.E4, enums.Capture),
			"rnbqkbnr/ppp1pppp/8/3P4/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2",
		},
		{
			"exd6",
			"rnbqkb1r/ppp2ppp/8/3pP3/3Qn3/5N2/PPP2PPP/RNB1KB1R w KQkq d6 0 6",
			bitboard.NewMove(enums.D6, enums.E5, enums.EPCapture),
			"rnbqkb1r/ppp2ppp/3P4/8/3Qn3/5N2/PPP2PPP/RNB1KB1R b KQkq - 0 6",
		},
		{
			"O-O",
			"r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 1 3",
			bitboard.NewMove(enums.G1, enums.E1, enums.KingCastle),
			"r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQ1RK1 b kq - 2 3",
		},
		{
			"O-O-O",
			"r3kb1r/ppp1qppp/2np1n2/1B2p3/P3P1b1/2NP1N2/1PP2PPP/R1BQ1RK1 b kq - 4 6",
			bitboard.NewMove(enums.C8, enums.E8, enums.QueenCastle),
			"2kr1b1r/ppp1qppp/2np1n2/1B2p3/P3P1b1/2NP1N2/1PP2PPP/R1BQ1RK1 w - - 5 7",
		},
		{
			"bxa8=N",
			"rnbqkbnr/pP3ppp/4p3/8/8/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			bitboard.NewMove(enums.A8, enums.B7, enums.KnightPromoCapture),
			"Nnbqkbnr/p4ppp/4p3/8/8/8/PPPP1PPP/RNBQKBNR b KQk - 0 1",
		},
		{
			"Bb5+",
			"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 1",
			bitboard.NewMove(enums.B5, enums.F1, enums.Quiet),
			"rnbqkbnr/ppp1pppp/8/1B1p4/4P3/8/PPPP1PPP/RNBQK1NR b KQkq - 1 1",
		},
	}
	for _, tc := range testcases {
		t.Logf("Passing test: %s\n", tc.san)
		game := NewGame(uuid.New(), fen.FEN2Bitboard(tc.beforeFEN), 180, 180)
		game.ProcessMove(tc.move)
		got := fen.Bitboard2FEN(game.Bitboard)
		if got != tc.expectedFEN {
			t.Fatalf("expected fen: %s, got: %s", tc.expectedFEN, got)
		}
		if game.Moves[len(game.Moves)-1].SAN != tc.san {
			t.Fatalf("expected san: %s, got: %s", tc.san, game.Moves[len(game.Moves)-1].SAN)
		}
	}
}

func BenchmarkProcessMove(b *testing.B) {
	game := NewGame(uuid.New(), nil, 180, 180)
	before := fen.Bitboard2FEN(game.Bitboard)
	for i := 0; i < b.N; i++ {
		game.ProcessMove(bitboard.NewMove(enums.E4, enums.E2, enums.DoublePawnPush))
		// Restore the game state.
		game.Bitboard = fen.FEN2Bitboard(before)
	}
}

func TestIsThreefoldRepetition(t *testing.T) {
	testcases := []struct {
		moves    []CompletedMove
		expected bool
	}{
		{[]CompletedMove{
			{0, "", "1kr5/Bb3R2/4p3/4Pn1p/R7/2P3p1/1KP4r/8 w - - 0 1", 0},
			{0, "", "k1r5/Bb3R2/4p3/4Pn1p/R7/2P3p1/1KP4r/8 w - - 0 1", 0},
			{0, "", "k1r5/1b3R2/4p3/4Pn1p/R7/2P3p1/1KP2B1r/8 w - - 0 1", 0},
			{0, "", "1kr5/1b3R2/4p3/4Pn1p/R7/2P3p1/1KP2B1r/8 w - - 0 1", 0},
			{0, "", "1kr5/Bb3R2/4p3/4Pn1p/R7/2P3p1/1KP4r/8 w - - 0 1", 0},
			{0, "", "k1r5/Bb3R2/4p3/4Pn1p/R7/2P3p1/1KP4r/8 w - - 0 1", 0},
			{0, "", "k1r5/1b3R2/4p3/4Pn1p/R7/2P3p1/1KP2B1r/8 w - - 0 1", 0},
		}, true},
		{[]CompletedMove{}, false},
	}
	for _, tc := range testcases {
		g := NewGame(uuid.New(), nil, 180, 0)
		g.Moves = tc.moves
		got := g.isThreefoldRepetition()
		if got != tc.expected {
			t.Fatalf("expected: %t, got: %t", tc.expected, got)
		}
	}
}

func TestIsInsufficientMaterial(t *testing.T) {
	testcases := []struct {
		fen      string
		pieces   [12]uint64
		expected bool
	}{
		{
			"8/8/1K4p1/7p/1P4k1/8/8/8 w - - 0 44",
			[12]uint64{
				0x2000000, 0x408000000000, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			false,
		},
		{
			"8/8/1K6/8/1P4k1/8/8/8 w - - 0 44",
			[12]uint64{
				0x20000000, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0,
			},
			false,
		},
		{
			"8/8/1K6/8/2N3k1/8/8/8 w - - 0 44",
			[12]uint64{
				0x0, 0x0, 0x4000000, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0,
			},
			true,
		},
		{
			"8/8/1K1N4/8/2N3k1/8/8/8 w - - 0 44",
			[12]uint64{
				0x0, 0x0, 0x8004000000, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0, 0x0, 0x0, 0x0,
			},
			false,
		},
		{
			"8/8/1K4b1/8/1B4k1/8/8/8 w - - 0 44",
			[12]uint64{
				0x0, 0x0, 0x0, 0x0, 0x2000000, 0x400000000000,
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			false,
		},
		{
			"8/5b2/1K6/8/1B4k1/8/8/8 w - - 0 44",
			[12]uint64{
				0x0, 0x0, 0x0, 0x0, 0x2000000, 0x40000000000000,
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			},
			true,
		},
	}
	for _, tc := range testcases {
		t.Logf("passing test: %s\n", tc.fen)
		g := NewGame(uuid.New(), bitboard.NewBitboard(tc.pieces, enums.White,
			[4]bool{false, false, false, false},
			enums.NoSquare, 0, 44), 180, 180)

		if tc.expected != g.isInsufficientMaterial() {
			t.Fatalf("expected: %t, got: %t\n", tc.expected, !tc.expected)
		}
	}
}
