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
				{File: enums.E, Rank: 3}: enums.PawnForward,
				{File: enums.E, Rank: 4}: enums.PawnForward,
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
				{File: enums.E, Rank: 3}: enums.PawnForward,
				{File: enums.E, Rank: 4}: enums.PawnForward,
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
				{File: enums.D, Rank: 9}: enums.PawnForward, // since the MoveCounter=0
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
				{File: enums.D, Rank: 6}: enums.PawnForward,
				{File: enums.D, Rank: 7}: enums.PawnForward, // since the MoveCounter=0
				{File: enums.E, Rank: 6}: enums.EnPassant,
			},
		},
		{
			"Black_Pawn",
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

func TestPawnMove(t *testing.T) {
	epPawn := pieces.NewPawn(enums.White, helpers.NewPos(enums.B, 4))
	epPawn.IsEnPassant = true
	bepPawn := pieces.NewPawn(enums.Black, helpers.NewPos(enums.H, 5))
	bepPawn.IsEnPassant = true

	testcases := []struct {
		name          string
		pawn          *pieces.Pawn
		pieces        map[helpers.Pos]pieces.Piece
		move          *helpers.Move
		expectedRes   bool
		expectedBoard map[helpers.Pos]pieces.Piece
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
		},
		{
			"illegal_move_c6-g7",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.C, 6)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.C, Rank: 6}: pieces.NewPawn(enums.White, helpers.NewPos(enums.C, 6)),
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.G, 7),
				From:             helpers.NewPos(enums.E, 2),
				PromotionPayload: 0,
			},
			false,
			map[helpers.Pos]pieces.Piece{
				{File: enums.C, Rank: 6}: pieces.NewPawn(enums.White, helpers.NewPos(enums.C, 6)),
			},
		},
		{
			"legal_capture_d5-e6",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 5)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.D, Rank: 5}: pieces.NewPawn(enums.White, helpers.NewPos(enums.D, 5)),
				{File: enums.E, Rank: 6}: pieces.NewKnight(enums.Black, helpers.NewPos(enums.E, 6)),
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.E, 6),
				From:             helpers.NewPos(enums.D, 5),
				PromotionPayload: 0,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 6}: pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 6)),
			},
		},
		{
			"black_en_passant_a4-b3",
			pieces.NewPawn(enums.Black, helpers.NewPos(enums.A, 4)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.A, Rank: 4}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.A, 4)),
				epPawn.Pos:               epPawn,
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.B, 3),
				From:             helpers.NewPos(enums.A, 4),
				PromotionPayload: 0,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.B, Rank: 3}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.B, 3)),
			},
		},
		{
			"white_en_passant_g5-h6",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.G, 5)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.G, Rank: 5}: pieces.NewPawn(enums.White, helpers.NewPos(enums.G, 5)),
				bepPawn.Pos:              bepPawn,
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.H, 6),
				From:             helpers.NewPos(enums.G, 5),
				PromotionPayload: 0,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.H, Rank: 6}: pieces.NewPawn(enums.White, helpers.NewPos(enums.H, 6)),
			},
		},
		{
			"white_promotion_e7-e8",
			pieces.NewPawn(enums.White, helpers.NewPos(enums.E, 7)),
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
		},
		{
			"black_capture_promotion_d2-c1",
			pieces.NewPawn(enums.Black, helpers.NewPos(enums.D, 2)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.D, Rank: 2}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.D, 2)),
				{File: enums.C, Rank: 1}: pieces.NewRook(enums.White, helpers.NewPos(enums.C, 1)),
			},
			&helpers.Move{
				To:               helpers.NewPos(enums.C, 1),
				From:             helpers.NewPos(enums.D, 2),
				PromotionPayload: enums.Rook,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.C, Rank: 1}: pieces.NewRook(enums.Black, helpers.NewPos(enums.C, 1)),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expectedRes := tc.expectedRes
			expectedBoard := tc.expectedBoard

			gotRes := tc.pawn.Move(tc.pieces, tc.move)
			gotBoard := tc.pieces

			if expectedRes != gotRes {
				t.Errorf("expected result: %t, got result: %t", expectedRes, gotRes)
			}
			for pos, piece := range expectedBoard {
				if gotBoard[pos].GetPosition() != piece.GetPosition() {
					t.Errorf("expected board: %v, got board: %v", expectedBoard, gotBoard)
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
			"all_possible_moves",
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
			"stalemate",
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
			"defended_piece",
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
			"battle_royale",
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

func TestKingMove(t *testing.T) {
	testcases := []struct {
		name          string
		king          *pieces.King
		pieces        map[helpers.Pos]pieces.Piece
		move          *helpers.Move
		expectedRes   bool
		expectedBoard map[helpers.Pos]pieces.Piece
	}{
		{
			"illegal_move_e6-e5",
			pieces.NewKing(enums.White, helpers.NewPos(enums.E, 6)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 6}: pieces.NewKing(enums.White, helpers.NewPos(enums.E, 6)),
				{File: enums.D, Rank: 6}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.D, 6)),
				{File: enums.F, Rank: 6}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.F, 6)),
			},
			&helpers.Move{
				To:   helpers.NewPos(enums.E, 5),
				From: helpers.NewPos(enums.E, 6),
			},
			false,
			map[helpers.Pos]pieces.Piece{
				{File: enums.E, Rank: 6}: pieces.NewKing(enums.White, helpers.NewPos(enums.E, 6)),
				{File: enums.D, Rank: 6}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.D, 6)),
				{File: enums.F, Rank: 6}: pieces.NewPawn(enums.Black, helpers.NewPos(enums.F, 6)),
			},
		},
		{
			"legal_move_b7-b6",
			pieces.NewKing(enums.Black, helpers.NewPos(enums.B, 7)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.B, Rank: 7}: pieces.NewKing(enums.Black, helpers.NewPos(enums.B, 7)),
				{File: enums.B, Rank: 6}: pieces.NewPawn(enums.White, helpers.NewPos(enums.B, 6)),
				{File: enums.A, Rank: 1}: pieces.NewRook(enums.White, helpers.NewPos(enums.A, 1)),
				{File: enums.H, Rank: 7}: pieces.NewRook(enums.White, helpers.NewPos(enums.H, 7)),
			},
			&helpers.Move{
				To:   helpers.NewPos(enums.B, 6),
				From: helpers.NewPos(enums.B, 7),
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.B, Rank: 6}: pieces.NewKing(enums.Black, helpers.NewPos(enums.B, 6)),
				{File: enums.A, Rank: 1}: pieces.NewRook(enums.White, helpers.NewPos(enums.A, 1)),
				{File: enums.H, Rank: 7}: pieces.NewRook(enums.White, helpers.NewPos(enums.H, 7)),
			},
		},
		{
			"black_0-0",
			pieces.NewKing(enums.Black, helpers.NewPos(enums.E, 8)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.H, Rank: 8}: pieces.NewRook(enums.Black, helpers.NewPos(enums.H, 8)),
			},
			&helpers.Move{
				To:       helpers.NewPos(enums.G, 8),
				From:     helpers.NewPos(enums.E, 8),
				MoveType: enums.ShortCastling,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.G, Rank: 8}: pieces.NewKing(enums.Black, helpers.NewPos(enums.G, 8)),
				{File: enums.F, Rank: 8}: pieces.NewRook(enums.Black, helpers.NewPos(enums.F, 8)),
			},
		},
		{
			"white_0-0-0",
			pieces.NewKing(enums.White, helpers.NewPos(enums.E, 1)),
			map[helpers.Pos]pieces.Piece{
				{File: enums.A, Rank: 1}: pieces.NewRook(enums.White, helpers.NewPos(enums.A, 1)),
			},
			&helpers.Move{
				To:       helpers.NewPos(enums.C, 1),
				From:     helpers.NewPos(enums.E, 1),
				MoveType: enums.LongCastling,
			},
			true,
			map[helpers.Pos]pieces.Piece{
				{File: enums.C, Rank: 1}: pieces.NewKing(enums.White, helpers.NewPos(enums.C, 1)),
				{File: enums.D, Rank: 1}: pieces.NewRook(enums.White, helpers.NewPos(enums.D, 1)),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expectedRes := tc.expectedRes
			expectedBoard := tc.expectedBoard

			gotRes := tc.king.Move(tc.pieces, tc.move)
			gotBoard := tc.pieces

			if expectedRes != gotRes {
				t.Errorf("expected result: %t, got result: %t", expectedRes, gotRes)
			}
			for pos, piece := range expectedBoard {
				if gotBoard[pos].GetPosition() != piece.GetPosition() {
					t.Errorf("expected board: %v, got board: %v", expectedBoard, gotBoard)
				}
			}
		})
	}
}
