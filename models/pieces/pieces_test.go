package pieces_test

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
	"chess-api/models/pieces"
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
			"First_Moves",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
			map[helpers.Pos]pieces.Piece{}, // empty board
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 3}: enums.Defend,
				{File: enums.F, Rank: 3}: enums.Defend,
				{File: enums.E, Rank: 3}: enums.Basic,
				{File: enums.E, Rank: 4}: enums.Basic,
			},
		},
		{
			"Capture",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.D, Rank: 3}: pieces.NewPawn(enums.Black,
					helpers.NewPos(enums.D, 3)),
			},
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 3}: enums.Basic,
				{File: enums.F, Rank: 3}: enums.Defend,
				{File: enums.E, Rank: 3}: enums.Basic,
				{File: enums.E, Rank: 4}: enums.Basic,
			},
		},
		{
			"Promotion",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 7)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 8}: pieces.NewPawn(enums.Black,
					helpers.NewPos(enums.E, 8)),
			},
			map[helpers.Pos]enums.MoveType{
				{File: enums.D, Rank: 8}: enums.Promotion,
				{File: enums.D, Rank: 9}: enums.Basic, // since the pawn hasnt moved
				{File: enums.E, Rank: 8}: enums.Promotion,
				{File: enums.C, Rank: 8}: enums.Defend,
			},
		},
		{
			"En_Passant",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 5)),
			map[helpers.Pos]pieces.Piece{
				epPawn.Pos: epPawn,
			},
			map[helpers.Pos]enums.MoveType{
				{File: enums.C, Rank: 6}: enums.Defend,
				{File: enums.D, Rank: 6}: enums.Basic,
				{File: enums.D, Rank: 7}: enums.Basic, // since the pawn hasnt moved
				{File: enums.E, Rank: 6}: enums.EnPassant,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected

			got := tc.pawn.GetPossibleMoves(tc.pieces)

			for pos := range expected {
				if got[pos] != expected[pos] {
					t.Error("got: ", got)
					t.Error("expected: ", expected)
				}
			}
		})
	}
}
