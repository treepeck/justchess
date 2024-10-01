package pieces_test

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
	"chess-api/models/pieces"
	"testing"
)

var enPassantPawn = pieces.NewPawn(enums.Black, helpers.Pos{
	File: enums.D, Rank: 5,
})

var testcases = []struct {
	name     string
	piece    pieces.Piece
	expected []helpers.Pos
	pieces   map[helpers.Pos]pieces.Piece
}{
	{
		name:  "White pawn d2",
		piece: pieces.NewPawn(enums.White, helpers.Pos{File: enums.D, Rank: 2}),
		expected: []helpers.Pos{
			{File: enums.D, Rank: 3},
			{File: enums.D, Rank: 4},
		},
		pieces: map[helpers.Pos]pieces.Piece{},
	},
	{
		name:     "White pawn d8",
		piece:    pieces.NewPawn(enums.White, helpers.Pos{File: enums.D, Rank: 8}),
		expected: []helpers.Pos{},
		pieces:   map[helpers.Pos]pieces.Piece{},
	},
	{
		name:  "Black pawn, can capture both sides",
		piece: pieces.NewPawn(enums.Black, helpers.Pos{File: enums.E, Rank: 7}),
		expected: []helpers.Pos{
			{File: enums.E, Rank: 6},
			{File: enums.E, Rank: 5},
			{File: enums.D, Rank: 6},
			{File: enums.F, Rank: 6},
		},
		pieces: map[helpers.Pos]pieces.Piece{
			{File: enums.D, Rank: 6}: pieces.NewPawn(enums.White, helpers.Pos{File: enums.D, Rank: 6}),
			{File: enums.F, Rank: 6}: pieces.NewPawn(enums.White, helpers.Pos{File: enums.F, Rank: 6}),
		},
	},
	{
		name:  "Black knight",
		piece: pieces.NewKnight(enums.Black, helpers.Pos{File: enums.D, Rank: 4}),
		expected: []helpers.Pos{
			{File: enums.F, Rank: 5},
			{File: enums.B, Rank: 5},
			{File: enums.B, Rank: 3},
			{File: enums.C, Rank: 6},
			{File: enums.E, Rank: 2},
			{File: enums.C, Rank: 2},
			{File: enums.E, Rank: 6},
		},
		pieces: map[helpers.Pos]pieces.Piece{
			{File: enums.F, Rank: 5}: pieces.NewPawn(enums.White, helpers.Pos{File: enums.F, Rank: 5}),
			{File: enums.F, Rank: 3}: pieces.NewPawn(enums.Black, helpers.Pos{File: enums.F, Rank: 3}),
		},
	},
	{
		name:  "White knight, end of board",
		piece: pieces.NewKnight(enums.White, helpers.Pos{File: enums.A, Rank: 8}),
		expected: []helpers.Pos{
			{File: enums.C, Rank: 7},
			{File: enums.B, Rank: 6},
		},
		pieces: map[helpers.Pos]pieces.Piece{
			{File: enums.C, Rank: 7}: pieces.NewPawn(enums.Black, helpers.Pos{File: enums.C, Rank: 7}),
		},
	},
	{
		name:     "White rook, start pos",
		piece:    pieces.NewRook(enums.White, helpers.Pos{File: enums.A, Rank: 1}),
		expected: []helpers.Pos{},
		pieces: map[helpers.Pos]pieces.Piece{
			{File: enums.A, Rank: 2}: pieces.NewPawn(enums.White, helpers.Pos{File: enums.A, Rank: 2}),
			{File: enums.B, Rank: 1}: pieces.NewKnight(enums.White, helpers.Pos{File: enums.B, Rank: 1}),
		},
	},
	{
		name:  "Black rook, middle of the board",
		piece: pieces.NewRook(enums.Black, helpers.Pos{File: enums.D, Rank: 4}),
		expected: []helpers.Pos{
			{File: enums.D, Rank: 3},
			{File: enums.D, Rank: 2},
			{File: enums.D, Rank: 1},
			{File: enums.D, Rank: 5},
			{File: enums.D, Rank: 6},
			{File: enums.C, Rank: 4},
			{File: enums.B, Rank: 4},
			{File: enums.A, Rank: 4},
			{File: enums.E, Rank: 4},
			{File: enums.F, Rank: 4},
			{File: enums.G, Rank: 4},
			{File: enums.H, Rank: 4},
		},
		pieces: map[helpers.Pos]pieces.Piece{
			{File: enums.D, Rank: 6}: pieces.NewKnight(enums.White, helpers.Pos{File: enums.D, Rank: 6}),
			{File: enums.D, Rank: 7}: pieces.NewPawn(enums.White, helpers.Pos{File: enums.D, Rank: 7}),
		},
	},
	{
		name:  "White bishop e5",
		piece: pieces.NewBishop(enums.White, helpers.Pos{File: enums.E, Rank: 5}),
		expected: []helpers.Pos{
			{File: enums.D, Rank: 6},
			{File: enums.C, Rank: 7},
			{File: enums.D, Rank: 4},
			{File: enums.C, Rank: 3},
			{File: enums.F, Rank: 6},
			{File: enums.G, Rank: 7},
			{File: enums.H, Rank: 8},
		},
		pieces: map[helpers.Pos]pieces.Piece{
			{File: enums.C, Rank: 7}: pieces.NewPawn(enums.Black, helpers.Pos{File: enums.C, Rank: 7}),
			{File: enums.B, Rank: 2}: pieces.NewPawn(enums.White, helpers.Pos{File: enums.B, Rank: 2}),
			{File: enums.F, Rank: 4}: pieces.NewQueen(enums.White, helpers.Pos{File: enums.F, Rank: 4}),
		},
	},
}

func TestGetAvailibleMoves(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected

			if tc.name == "White pawn, en passant" {
				tc.piece.(*pieces.Pawn).MovesCounter = 2
				enPassantPawn.MovesCounter = 1
			}

			got := tc.piece.GetAvailibleMoves(tc.pieces)

			if len(expected) != len(got) {
				t.Error("different lengths", expected, got)
			} else {
				for i := range got {
					if expected[i] != got[i] {
						t.Error(tc.name, expected, got)
					}
				}
			}
		})
	}
}
