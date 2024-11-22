package helpers_test

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"testing"
)

func TestToLAN(t *testing.T) {
	testcases := []struct {
		name      string
		move      helpers.Move
		pieceType enums.PieceType
	}{
		{
			"♖d1-a1",
			helpers.Move{
				To:       helpers.NewPos(enums.A, 1),
				From:     helpers.NewPos(enums.D, 1),
				MoveType: enums.Basic,
			},
			enums.Rook,
		},
		{
			"♗h1xb7+",
			helpers.Move{
				To:        helpers.NewPos(enums.B, 7),
				From:      helpers.NewPos(enums.H, 1),
				MoveType:  enums.Basic,
				IsCapture: true,
				IsCheck:   true,
			},
			enums.Bishop,
		},
		{
			"g7xf8=♕",
			helpers.Move{
				To:               helpers.NewPos(enums.F, 8),
				From:             helpers.NewPos(enums.G, 7),
				MoveType:         enums.Promotion,
				IsCapture:        true,
				PromotionPayload: enums.Queen,
			},
			enums.Pawn,
		},
		{
			"0-0-0#",
			helpers.Move{
				To:          helpers.NewPos(enums.A, 8),
				From:        helpers.NewPos(enums.E, 8),
				MoveType:    enums.LongCastling,
				IsCheck:     true,
				IsCheckmate: true,
			},
			enums.King,
		},
		{
			"0-0+",
			helpers.Move{
				To:       helpers.NewPos(enums.H, 8),
				From:     helpers.NewPos(enums.E, 8),
				MoveType: enums.ShortCastling,
				IsCheck:  true,
			},
			enums.King,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			lan := tc.move.ToLAN(tc.pieceType)

			if lan != tc.name {
				t.Errorf("expected LAN: %s, got: %s", tc.name, lan)
			}
		})
	}
}
