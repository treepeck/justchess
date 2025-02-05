package game

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"justchess/pkg/game/fen"
	"testing"
)

var dummyGame = NewGame(0, bitboard.NewBitboard([12]uint64{
	0x1000EF00, 0x0, 0x40040, 0x0, 0x24, 0x0,
	0x81, 0x0, 0x8, 0x80000000, 0x10, 0x0},
	enums.White, [4]bool{true, true, true, true},
	-1, 0, 0), 180, 0)

func TestProcessMove(t *testing.T) {
	testcases := []struct {
		game              *Game
		move              bitboard.Move
		expResult         enums.Result
		expPieces         [12]uint64
		expActiveColor    enums.Color
		expCastlingRights [4]bool
		expEPTarget       int
		expHalfmoveCnt    int
		expFullmoveCnt    int
	}{
		{
			dummyGame,
			bitboard.NewMove(enums.H5, enums.D1, enums.Quiet),
			enums.Unknown,
			[12]uint64{
				0x1000EF00, 0x0, 0x40040, 0x0, 0x24, 0x0,
				0x81, 0x0, 1 << 39, 0x80000000, 0x10, 0x0,
			},
			enums.Black,
			[4]bool{true, true, true, true},
			-1,
			1,
			0,
		},
	}
	for _, tc := range testcases {
		tc.game.ProcessMove(tc.move)
		if tc.game.Result != tc.expResult {
			t.Fatalf("expected result: %d, got: %d", tc.expResult, tc.game.Result)
		}
		for i, bb := range tc.game.Bitboard.Pieces {
			if bb != tc.expPieces[i] {
				t.Fatalf("expected pieces: %v, got: %v", tc.expPieces, tc.game.Bitboard.Pieces)
			}
		}
		if tc.game.Bitboard.ActiveColor != tc.expActiveColor {
			t.Fatalf("expected color: %d, got: %d", tc.expActiveColor, tc.game.Bitboard.ActiveColor)
		}
		for i, r := range tc.game.Bitboard.CastlingRights {
			if r != tc.expCastlingRights[i] {
				t.Fatalf("expected rights: %v, got: %v", tc.expCastlingRights, tc.game.Bitboard.CastlingRights)
			}
		}
		if tc.expEPTarget != tc.game.Bitboard.EPTarget {
			t.Fatalf("expected rights: %v, got: %v", tc.expCastlingRights, tc.game.Bitboard.CastlingRights)
		}
	}
}

// TODO: this has some weird behaviour.
func BenchmarkProcessMove(b *testing.B) {
	FENBefore := fen.Bitboard2FEN(dummyGame.Bitboard)
	for i := 0; i < b.N; i++ {
		dummyGame.Bitboard = fen.FEN2Bitboard(FENBefore)
		dummyGame.ProcessMove(bitboard.NewMove(enums.H5, enums.D1, enums.Quiet))
	}
}

func TestIsThreefoldRepetition(t *testing.T) {
	testcases := []struct {
		moves    []CompletedMove
		expected bool
	}{
		{[]CompletedMove{
			{"", "1kr5/Bb3R2/4p3/4Pn1p/R7/2P3p1/1KP4r/8 w - - 0 1"},
			{"", "k1r5/Bb3R2/4p3/4Pn1p/R7/2P3p1/1KP4r/8 w - - 0 1"},
			{"", "k1r5/1b3R2/4p3/4Pn1p/R7/2P3p1/1KP2B1r/8 w - - 0 1"},
			{"", "1kr5/1b3R2/4p3/4Pn1p/R7/2P3p1/1KP2B1r/8 w - - 0 1"},
			{"", "1kr5/Bb3R2/4p3/4Pn1p/R7/2P3p1/1KP4r/8 w - - 0 1"},
			{"", "k1r5/Bb3R2/4p3/4Pn1p/R7/2P3p1/1KP4r/8 w - - 0 1"},
			{"", "k1r5/1b3R2/4p3/4Pn1p/R7/2P3p1/1KP2B1r/8 w - - 0 1"},
		}, true},
		{[]CompletedMove{}, false},
	}
	for _, tc := range testcases {
		g := NewGame(0, nil, 180, 0)
		g.Moves = tc.moves
		got := g.isThreefoldRepetition()
		if got != tc.expected {
			t.Fatalf("expected: %t, got: %t", tc.expected, got)
		}
	}
}
