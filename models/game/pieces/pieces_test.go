package pieces_test

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"chess-api/models/game/pieces"
	"testing"
)

func TestPawnGetPossibleMoves(t *testing.T) {
	epPawn := pieces.NewPawn(enums.Black, helpers.NewPos(enums.E, 5))
	epPawn.IsEnPassant = true

	testcases := []struct {
		name     string
		pawn     *pieces.Pawn
		pieces   map[helpers.Pos]pieces.Piece
		expected map[helpers.Pos]enums.MoveType
	}{
		{
			"white_pawn_e2",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
			map[helpers.Pos]pieces.Piece{}, // empty board
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 3}: enums.Defend,
				{File: enums.F, Rank: 3}: enums.Defend,
				{File: enums.E, Rank: 3}: enums.PawnForward,
				{File: enums.E, Rank: 4}: enums.PawnForward,
			},
		},
		{
			"white_pawn_e2_capture",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.D, Rank: 3}: pieces.NewPawn(enums.Black,
					helpers.NewPos(enums.D, 3)),
			},
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 3}: enums.Basic,
				{File: enums.F, Rank: 3}: enums.Defend,
				{File: enums.E, Rank: 3}: enums.PawnForward,
				{File: enums.E, Rank: 4}: enums.PawnForward,
			},
		},
		{
			"white_pawn_d7",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 7)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 8}: pieces.NewPawn(enums.Black,
					helpers.NewPos(enums.E, 8)),
			},
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 8}: enums.Promotion,
				{File: enums.D, Rank: 9}: enums.PawnForward, // since the MoveCounter=0
				{File: enums.E, Rank: 8}: enums.Promotion,
				{File: enums.C, Rank: 8}: enums.Defend,
			},
		},
		{
			"white_pawn_d5_enPassant",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 5)),
			map[helpers.Pos]pieces.Piece{
				epPawn.Pos: epPawn,
			},
			map[helpers.Pos]enums.MoveType{
				{File: enums.C, Rank: 6}: enums.Defend,
				{File: enums.D, Rank: 6}: enums.PawnForward,
				{File: enums.D, Rank: 7}: enums.PawnForward, // since the MoveCounter=0
				{File: enums.E, Rank: 6}: enums.EnPassant,
			},
		},
		{
			"black_pawn_g3",
			pieces.NewPawn(enums.Black, helpers.NewPos(enums.G, 3)),
			map[helpers.Pos]pieces.Piece{},
			map[helpers.Pos]enums.MoveType{
				{File: enums.F, Rank: 2}: enums.Defend,
				{File: enums.H, Rank: 2}: enums.Defend,
				{File: enums.G, Rank: 2}: enums.PawnForward,
				{File: enums.G, Rank: 1}: enums.PawnForward, // since the MoveCounter=0
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected

			got := tc.pawn.GetPossibleMoves(tc.pieces)

			for pos := range expected {
				if got[pos] != expected[pos] {
					t.Errorf("expected: %v, got: %v", expected, got)
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
		expected map[helpers.Pos]enums.MoveType
	}{
		{
			"white_king_e5",
			pieces.NewKing(enums.White, helpers.NewPos(enums.E, 5)),
			map[helpers.Pos]pieces.Piece{}, // empty board
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 6}: enums.Basic,
				{File: enums.E, Rank: 6}: enums.Basic,
				{File: enums.F, Rank: 6}: enums.Basic,
				{File: enums.D, Rank: 5}: enums.Basic,
				{File: enums.F, Rank: 5}: enums.Basic,
				{File: enums.D, Rank: 4}: enums.Basic,
				{File: enums.E, Rank: 4}: enums.Basic,
				{File: enums.F, Rank: 4}: enums.Basic,
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
			map[helpers.Pos]enums.MoveType{},
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
			map[helpers.Pos]enums.MoveType{},
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
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 6}: enums.Basic,
				{File: enums.E, Rank: 6}: enums.Defend,
				{File: enums.F, Rank: 6}: enums.Basic,
			},
		},
		{
			"white_0-0",
			pieces.NewKing(enums.White, helpers.NewPos(enums.E, 1)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.H, Rank: 1}: pieces.NewRook(enums.White,
					helpers.NewPos(enums.H, 1)),
			},
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 1}: enums.Basic,
				{File: enums.D, Rank: 2}: enums.Basic,
				{File: enums.E, Rank: 2}: enums.Basic,
				{File: enums.F, Rank: 2}: enums.Basic,
				{File: enums.F, Rank: 1}: enums.Basic,
				{File: enums.G, Rank: 1}: enums.ShortCastling,
			},
		},
		{
			"black_0-0-0",
			pieces.NewKing(enums.Black, helpers.NewPos(enums.E, 8)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.A, Rank: 8}: pieces.NewRook(enums.Black,
					helpers.NewPos(enums.A, 8)),
			},
			map[helpers.Pos]enums.MoveType{
				{File: enums.F, Rank: 8}: enums.Basic,
				{File: enums.F, Rank: 7}: enums.Basic,
				{File: enums.E, Rank: 7}: enums.Basic,
				{File: enums.D, Rank: 7}: enums.Basic,
				{File: enums.D, Rank: 8}: enums.Basic,
				{File: enums.C, Rank: 8}: enums.LongCastling,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected

			got := tc.king.GetPossibleMoves(tc.pieces)

			for pos := range expected {
				if got[pos] != expected[pos] {
					t.Errorf("expected: %v, got: %v", expected, got)
				}
			}
		})
	}
}
