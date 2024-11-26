package pieces_test

import (
	"testing"

	"justchess/pkg/models/game/enums"
	"justchess/pkg/models/game/helpers"
	"justchess/pkg/models/game/pieces"
)

func TestPawnGetPossibleMoves(t *testing.T) {
	epPawn := pieces.NewPawn(enums.Black, helpers.NewPos(enums.E, 5))
	epPawn.IsEnPassant = true

	testcases := []struct {
		name     string
		pawn     *pieces.Pawn
		pieces   map[helpers.Pos]pieces.Piece
		expected []helpers.PossibleMove
	}{
		{
			"white_pawn_e2",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
			map[helpers.Pos]pieces.Piece{}, // empty board
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.E, 3), MoveType: enums.PawnForward},
				{To: helpers.NewPos(enums.E, 4), MoveType: enums.PawnForward},
				{To: helpers.NewPos(enums.D, 3), MoveType: enums.Defend},
				{To: helpers.NewPos(enums.F, 3), MoveType: enums.Defend},
			},
		},
		{
			"white_pawn_e2_capture",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.D, Rank: 3}: pieces.NewPawn(enums.Black,
					helpers.NewPos(enums.D, 3)),
			},
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.E, 3), MoveType: enums.PawnForward},
				{To: helpers.NewPos(enums.E, 4), MoveType: enums.PawnForward},
				{To: helpers.NewPos(enums.D, 3), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 3), MoveType: enums.Defend},
			},
		},
		{
			"white_pawn_d7",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 7)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 8}: pieces.NewPawn(enums.Black,
					helpers.NewPos(enums.E, 8)),
			},
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.D, 8), MoveType: enums.Promotion},
				{To: helpers.NewPos(enums.D, 9), MoveType: enums.PawnForward}, // since the MovesCounter=0
				{To: helpers.NewPos(enums.C, 8), MoveType: enums.Defend},
				{To: helpers.NewPos(enums.E, 8), MoveType: enums.Promotion},
			},
		},
		{
			"white_pawn_d5_enPassant",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 5)),
			map[helpers.Pos]pieces.Piece{
				epPawn.Pos: epPawn,
			},
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.D, 6), MoveType: enums.PawnForward},
				{To: helpers.NewPos(enums.D, 7), MoveType: enums.PawnForward},
				{To: helpers.NewPos(enums.C, 6), MoveType: enums.Defend},
				{To: helpers.NewPos(enums.E, 6), MoveType: enums.EnPassant},
			},
		},
		{
			"black_pawn_g3",
			pieces.NewPawn(enums.Black, helpers.NewPos(enums.G, 3)),
			map[helpers.Pos]pieces.Piece{},
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.G, 2), MoveType: enums.PawnForward},
				{To: helpers.NewPos(enums.G, 1), MoveType: enums.PawnForward},
				{To: helpers.NewPos(enums.F, 2), MoveType: enums.Defend},
				{To: helpers.NewPos(enums.H, 2), MoveType: enums.Defend},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pawn.GetPossibleMoves(tc.pieces)

			if len(got) != len(tc.expected) {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}

			for ind, pos := range tc.expected {
				if got[ind] != pos {
					t.Fatalf("expected: %v, got: %v", pos, got[ind])
				}
			}
		})
	}
}

func TestKingGetPossibleMoves(t *testing.T) {
	testcases := []struct {
		name     string
		king     *pieces.King
		pieces   map[helpers.Pos]pieces.Piece
		expected []helpers.PossibleMove
	}{
		{
			"white_king_e5",
			pieces.NewKing(enums.White, helpers.NewPos(enums.E, 5)),
			map[helpers.Pos]pieces.Piece{}, // empty board
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.D, 6), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.E, 6), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 6), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.D, 5), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 5), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.D, 4), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.E, 4), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 4), MoveType: enums.Basic},
			},
		},
		{
			"white_king_a1",
			pieces.NewKing(enums.White, helpers.NewPos(enums.A, 1)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.B, Rank: 8}: pieces.NewQueen(enums.Black,
					helpers.NewPos(enums.B, 8)),
				{File: enums.H, Rank: 2}: pieces.NewRook(enums.Black,
					helpers.NewPos(enums.H, 2)),
			},
			make([]helpers.PossibleMove, 0),
		},
		{
			"black_king_a8",
			pieces.NewKing(enums.Black, helpers.NewPos(enums.A, 8)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.B, Rank: 1}: pieces.NewRook(enums.White,
					helpers.NewPos(enums.B, 1)),
				{File: enums.H, Rank: 7}: pieces.NewRook(enums.White,
					helpers.NewPos(enums.H, 7)),
				{File: enums.A, Rank: 7}: pieces.NewKnight(enums.White,
					helpers.NewPos(enums.A, 7)),
			},
			make([]helpers.PossibleMove, 0),
		},
		{
			"black_king_e5",
			pieces.NewKing(enums.Black, helpers.NewPos(enums.E, 5)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 6}: pieces.NewPawn(enums.Black,
					helpers.NewPos(enums.E, 6)),
				{File: enums.E, Rank: 4}: pieces.NewPawn(enums.White,
					helpers.NewPos(enums.E, 4)),
				{File: enums.E, Rank: 3}: pieces.NewKing(enums.White,
					helpers.NewPos(enums.E, 3)),
			},
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.D, 6), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 6), MoveType: enums.Basic},
			},
		},
		{
			"white_0-0",
			pieces.NewKing(enums.White, helpers.NewPos(enums.E, 1)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.H, Rank: 1}: pieces.NewRook(enums.White,
					helpers.NewPos(enums.H, 1)),
			},
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.D, 2), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.E, 2), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 2), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.D, 1), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 1), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.G, 1), MoveType: enums.ShortCastling},
			},
		},
		{
			"black_0-0-0",
			pieces.NewKing(enums.Black, helpers.NewPos(enums.E, 8)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.A, Rank: 8}: pieces.NewRook(enums.Black,
					helpers.NewPos(enums.A, 8)),
			},
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.D, 8), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 8), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.D, 7), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.E, 7), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 7), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.C, 8), MoveType: enums.LongCastling},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.king.GetPossibleMoves(tc.pieces)

			if len(got) != len(tc.expected) {
				t.Fatalf("expected: %v, got: %v", tc.expected, got)
			}

			for pos := range tc.expected {
				if got[pos] != tc.expected[pos] {
					t.Fatalf("expected: %v, got: %v", tc.expected, got)
				}
			}
		})
	}
}

func TestKnightGetPossibleMoves(t *testing.T) {
	testcases := []struct {
		name     string
		knight   *pieces.Knight
		pieces   map[helpers.Pos]pieces.Piece
		expected []helpers.PossibleMove
	}{
		{
			"white_knight_e6",
			pieces.NewKnight(enums.White, helpers.NewPos(enums.E, 6)),
			make(map[helpers.Pos]pieces.Piece),
			[]helpers.PossibleMove{
				{To: helpers.NewPos(enums.G, 7), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.G, 5), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.C, 7), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.C, 5), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.D, 8), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 4), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.D, 4), MoveType: enums.Basic},
				{To: helpers.NewPos(enums.F, 8), MoveType: enums.Basic},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.knight.GetPossibleMoves(tc.pieces)

			if len(got) != len(tc.expected) {
				t.Errorf("expected: %v, got: %v", tc.expected, got)
			}

			for ind, pos := range tc.expected {
				if got[ind] != pos {
					t.Errorf("expected: %v, got: %v", pos, got[ind])
				}
			}
		})
	}
}
