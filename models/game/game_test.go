package game

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"chess-api/models/game/pieces"
	"testing"

	"github.com/google/uuid"
)

func TestGetPlayerValidMoves(t *testing.T) {
	g := NewG(uuid.Nil, enums.Blitz, 0, uuid.Nil, uuid.Nil)

	testcases := []struct {
		name        string
		pieces      map[helpers.Pos]pieces.Piece
		expectedVM  map[helpers.PossibleMove]bool
		currentTurn enums.Color
	}{
		{
			"8/8/8/1k6/R2q4/8/2R5/3K4 w - - 0 1",
			map[helpers.Pos]pieces.Piece{
				{File: enums.D, Rank: 1}: pieces.NewKing(enums.White, helpers.NewPos(enums.D, 1)),
				{File: enums.B, Rank: 5}: pieces.NewKing(enums.Black, helpers.NewPos(enums.B, 5)),
				{File: enums.C, Rank: 2}: pieces.NewRook(enums.White, helpers.NewPos(enums.C, 2)),
				{File: enums.A, Rank: 4}: pieces.NewRook(enums.White, helpers.NewPos(enums.A, 4)),
				{File: enums.D, Rank: 4}: pieces.NewQueen(enums.Black, helpers.NewPos(enums.D, 4)),
			},
			map[helpers.PossibleMove]bool{
				{
					To:        helpers.NewPos(enums.D, 2),
					From:      helpers.NewPos(enums.C, 2),
					MoveType:  enums.Basic,
					PieceType: enums.Rook,
				}: true,
				{
					To:        helpers.NewPos(enums.D, 4),
					From:      helpers.NewPos(enums.A, 4),
					MoveType:  enums.Basic,
					PieceType: enums.Rook,
				}: true,
				{
					To:        helpers.NewPos(enums.C, 1),
					From:      helpers.NewPos(enums.D, 1),
					MoveType:  enums.Basic,
					PieceType: enums.King,
				}: true,
				{
					To:        helpers.NewPos(enums.E, 2),
					From:      helpers.NewPos(enums.D, 1),
					MoveType:  enums.Basic,
					PieceType: enums.King,
				}: true,
				{
					To:        helpers.NewPos(enums.E, 1),
					From:      helpers.NewPos(enums.D, 1),
					MoveType:  enums.Basic,
					PieceType: enums.King,
				}: true,
			},
			enums.White,
		},
		{
			"4K3/8/8/8/7Q/8/3R1p2/2Rqk3 w - - 0 1",
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 1}: pieces.NewKing(enums.Black, helpers.NewPos(enums.E, 1)),
				{File: enums.E, Rank: 8}: pieces.NewKing(enums.White, helpers.NewPos(enums.E, 8)),
				{File: enums.C, Rank: 1}: pieces.NewRook(enums.White, helpers.NewPos(enums.C, 1)),
				{File: enums.D, Rank: 2}: pieces.NewRook(enums.White, helpers.NewPos(enums.D, 2)),
				{File: enums.F, Rank: 2}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.F, 2)),
				{File: enums.H, Rank: 4}: pieces.NewQueen(enums.White, helpers.NewPos(enums.H, 4)),
				{File: enums.D, Rank: 1}: pieces.NewQueen(enums.Black, helpers.NewPos(enums.D, 1)),
			},
			map[helpers.PossibleMove]bool{
				{
					To:        helpers.NewPos(enums.F, 1),
					From:      helpers.NewPos(enums.E, 1),
					MoveType:  enums.Basic,
					PieceType: enums.King,
				}: true,
				{
					To:        helpers.NewPos(enums.C, 1),
					From:      helpers.NewPos(enums.D, 1),
					MoveType:  enums.Basic,
					PieceType: enums.Queen,
				}: true,
				{
					To:        helpers.NewPos(enums.D, 2),
					From:      helpers.NewPos(enums.E, 1),
					MoveType:  enums.Basic,
					PieceType: enums.King,
				}: true,
			},
			enums.Black,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			g.Pieces = tc.pieces
			got := g.getPlayerValidMoves(tc.currentTurn)

			if len(got) != len(tc.expectedVM) {
				t.Errorf("expected len: %d, got: %d", len(tc.expectedVM), len(got))
				t.Errorf("expected: %v, got: %v", tc.expectedVM, got)
			}

			for pos := range tc.expectedVM {
				if got[pos] != tc.expectedVM[pos] {
					t.Errorf("expected: %v, got: %v", tc.expectedVM, got)
				}
			}
		})
	}
}

func TestHandleMove(t *testing.T) {
	g := NewG(uuid.Nil, enums.Blitz, 0, uuid.Nil, uuid.Nil)

	testcases := []struct {
		name           string
		piece          pieces.Piece
		pieces         map[helpers.Pos]pieces.Piece
		move           *helpers.Move
		expectedRes    bool
		expectedPieces map[helpers.Pos]pieces.Piece
		currentTurn    enums.Color
	}{
		{
			"legal_move_e2-e4",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 2}: pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.E, 4),
				From:             helpers.NewPos(enums.E, 2),
				PromotionPayload: 0,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 4}: pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 4)),
			},
			enums.White,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			g.Pieces = tc.pieces

			g.cvm = g.getPlayerValidMoves(tc.currentTurn)
			gotRes := g.HandleMove(*tc.move)
			gotBoard := g.Pieces

			if tc.expectedRes != gotRes {
				t.Errorf("expected result: %t, got result: %t", tc.expectedRes, gotRes)
			}
			for pos, piece := range tc.expectedPieces {
				if gotBoard[pos].GetPosition() != piece.GetPosition() {
					t.Errorf("expected board: %v, got board: %v", tc.expectedPieces, gotBoard)
				}
			}
		})
	}
}
