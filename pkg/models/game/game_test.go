package game

import (
	"testing"

	"justchess/pkg/models/game/enums"
	"justchess/pkg/models/game/helpers"
	"justchess/pkg/models/game/pieces"

	"github.com/google/uuid"
)

func TestGetPlayerValidMoves(t *testing.T) {
	g := NewG(uuid.Nil, enums.Blitz, 0, uuid.Nil, uuid.Nil)

	testcases := []struct {
		name        string
		pieces      map[helpers.Pos]pieces.Piece
		expectedVM  map[helpers.Pos][]helpers.PossibleMove
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
			map[helpers.Pos][]helpers.PossibleMove{
				{File: enums.C, Rank: 2}: {
					helpers.NewPM(helpers.NewPos(enums.D, 2), enums.Basic),
				},
				{File: enums.A, Rank: 4}: {
					helpers.NewPM(helpers.NewPos(enums.D, 4), enums.Basic),
				},
				{File: enums.D, Rank: 1}: {
					helpers.NewPM(helpers.NewPos(enums.E, 2), enums.Basic),
					helpers.NewPM(helpers.NewPos(enums.C, 1), enums.Basic),
					helpers.NewPM(helpers.NewPos(enums.E, 1), enums.Basic),
				},
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
			map[helpers.Pos][]helpers.PossibleMove{
				{File: enums.E, Rank: 1}: {
					helpers.NewPM(helpers.NewPos(enums.D, 2), enums.Basic),
					helpers.NewPM(helpers.NewPos(enums.F, 1), enums.Basic),
				},
				{File: enums.D, Rank: 1}: {
					helpers.NewPM(helpers.NewPos(enums.C, 1), enums.Basic),
				},
			},
			enums.Black,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			g.Pieces = tc.pieces
			g.CurrentTurn = tc.currentTurn
			got := g.getValidMoves(g.CurrentTurn)

			if len(got) != len(tc.expectedVM) {
				t.Fatalf("expected: %v, got: %v", tc.expectedVM, got)
			}

			for pos, pm := range tc.expectedVM {
				for ind, m := range pm {
					p := got[pos][ind]
					if !p.To.IsEqual(m.To) || p.MoveType != m.MoveType {
						t.Fatalf("expected: %v, got: %v", tc.expectedVM, got)
					}
				}
			}
		})
	}
}

func TestHandleMove(t *testing.T) {
	g := NewG(uuid.Nil, enums.Blitz, 0, uuid.Nil, uuid.Nil)

	testcases := []struct {
		name           string
		pieces         map[helpers.Pos]pieces.Piece
		move           *helpers.Move
		expectedRes    bool
		expectedPieces map[helpers.Pos]pieces.Piece
		currentTurn    enums.Color
	}{
		{
			"legal_move_e2-e4",
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 2}: pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 2)),
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
		{
			"legal_white_0-0",
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 1}: pieces.NewKing(enums.White, helpers.NewPos(enums.E, 1)),
				{File: enums.H, Rank: 1}: pieces.NewRook(enums.White, helpers.NewPos(enums.H,
					1)),
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.G, 1),
				From:             helpers.NewPos(enums.E, 1),
				PromotionPayload: 0,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.G, Rank: 1}: pieces.NewKing(enums.White, helpers.NewPos(enums.G, 1)),
				{File: enums.F, Rank: 1}: pieces.NewRook(enums.White, helpers.NewPos(enums.F, 1)),
			},
			enums.White,
		},
		{
			"legal_black_0-0-0",
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 8}: pieces.NewKing(enums.Black, helpers.NewPos(enums.E, 8)),
				{File: enums.A, Rank: 8}: pieces.NewRook(enums.Black, helpers.NewPos(enums.A,
					8)),
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.C, 8),
				From:             helpers.NewPos(enums.E, 8),
				PromotionPayload: 0,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.C, Rank: 8}: pieces.NewKing(enums.Black, helpers.NewPos(enums.C, 8)),
				{File: enums.D, Rank: 8}: pieces.NewRook(enums.Black, helpers.NewPos(enums.D, 8)),
			},
			enums.Black,
		},
		{
			"e8=Q",
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 7}: pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 7)),
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.E, 8),
				From:             helpers.NewPos(enums.E, 7),
				PromotionPayload: enums.Queen,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 8}: pieces.NewQueen(enums.White, helpers.NewPos(enums.E, 8)),
			},
			enums.White,
		},
		{
			"exd6_en_passant",
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 5}: pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 5)),
				{File: enums.D, Rank: 5}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.D, 5)),
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.D, 6),
				From:             helpers.NewPos(enums.E, 5),
				PromotionPayload: 0,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.D, Rank: 6}: pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 6)),
			},
			enums.White,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			g.Pieces = tc.pieces
			g.CurrentTurn = tc.currentTurn

			if tc.name == "exd6_en_passant" {
				g.Pieces[helpers.NewPos(enums.D, 5)].(*pieces.Pawn).IsEnPassant = true
			}

			g.CurrentValidMoves = g.getValidMoves(g.CurrentTurn)
			gotRes := g.HandleMove(tc.move)
			gotBoard := g.Pieces

			if tc.expectedRes != gotRes {
				t.Errorf("expected result: %t, got result: %t", tc.expectedRes, gotRes)
			}
			if len(tc.expectedPieces) != len(gotBoard) {
				t.Errorf("expected board: %v, got board: %v", tc.expectedPieces, gotBoard)
			}
			for pos, piece := range tc.expectedPieces {
				if gotBoard[pos].GetPosition() != piece.GetPosition() {
					t.Errorf("expected board: %v, got board: %v", tc.expectedPieces, gotBoard)
				}
			}
		})
	}
}

func TestToFEN(t *testing.T) {
	testcases := []struct {
		name   string
		pieces map[helpers.Pos]pieces.Piece
	}{
		{
			"8/8/4p3/8/2k5/8/8/8",
			map[helpers.Pos]pieces.Piece{
				{File: enums.C, Rank: 4}: pieces.NewKing(enums.Black, helpers.NewPos(enums.C, 4)),
				{File: enums.E, Rank: 6}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.E, 6)),
			},
		},
		{
			"r1bk3r/p2pBpNp/n4n2/1p1NP2P/6P1/3P4/P1P1K3/q5b1",
			map[helpers.Pos]pieces.Piece{
				{File: enums.A, Rank: 8}: pieces.NewRook(enums.Black, helpers.NewPos(enums.A, 8)),
				{File: enums.C, Rank: 8}: pieces.NewBishop(enums.Black, helpers.NewPos(enums.C, 8)),
				{File: enums.D, Rank: 8}: pieces.NewKing(enums.Black, helpers.NewPos(enums.D, 8)),
				{File: enums.H, Rank: 8}: pieces.NewRook(enums.Black, helpers.NewPos(enums.H, 8)),
				{File: enums.A, Rank: 7}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.A, 7)),
				{File: enums.D, Rank: 7}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.D, 7)),
				{File: enums.E, Rank: 7}: pieces.NewBishop(enums.White, helpers.NewPos(enums.E, 7)),
				{File: enums.F, Rank: 7}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.F, 7)),
				{File: enums.G, Rank: 7}: pieces.NewKnight(enums.White, helpers.NewPos(enums.G, 7)),
				{File: enums.H, Rank: 7}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.H, 7)),
				{File: enums.A, Rank: 6}: pieces.NewKnight(enums.Black, helpers.NewPos(enums.A, 6)),
				{File: enums.F, Rank: 6}: pieces.NewKnight(enums.Black, helpers.NewPos(enums.F, 6)),
				{File: enums.B, Rank: 5}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.B, 5)),
				{File: enums.D, Rank: 5}: pieces.NewKnight(enums.White, helpers.NewPos(enums.D, 5)),
				{File: enums.E, Rank: 5}: pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 5)),
				{File: enums.H, Rank: 5}: pieces.NewPawn(enums.White, helpers.NewPos(enums.H, 5)),
				{File: enums.G, Rank: 4}: pieces.NewPawn(enums.White, helpers.NewPos(enums.G, 4)),
				{File: enums.D, Rank: 3}: pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 3)),
				{File: enums.A, Rank: 2}: pieces.NewPawn(enums.White, helpers.NewPos(enums.A, 2)),
				{File: enums.C, Rank: 2}: pieces.NewPawn(enums.White, helpers.NewPos(enums.C, 2)),
				{File: enums.E, Rank: 2}: pieces.NewKing(enums.White, helpers.NewPos(enums.E, 2)),
				{File: enums.A, Rank: 1}: pieces.NewQueen(enums.Black, helpers.NewPos(enums.A, 1)),
				{File: enums.G, Rank: 1}: pieces.NewBishop(enums.Black, helpers.NewPos(enums.G, 1)),
			},
		},
	}

	for _, tc := range testcases {
		g := NewG(uuid.Nil, enums.Blitz, 0, uuid.Nil, uuid.Nil)
		g.Pieces = tc.pieces
		got := g.ToFEN()

		if got != tc.name {
			t.Fatalf("expected: %s, got: %s", tc.name, got)
		}
	}
}
