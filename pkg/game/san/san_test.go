package san

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"justchess/pkg/game/fen"
	"testing"
)

func TestMove2SAN(t *testing.T) {
	testcases := []struct {
		move     bitboard.Move
		bb       *bitboard.Bitboard
		pt       enums.PieceType
		expected string
	}{
		{
			bitboard.NewMove(enums.E2, enums.C3, enums.Quiet),
			fen.FEN2Bitboard("8/8/8/8/8/2N5/8/4K1N1 w - - 0 1"),
			enums.WhiteKnight,
			"Nce2",
		},
		{
			// Similar case to the previous one, except the knight c3 is pinned by the black
			// bishop, so the disambiguation is not needed.
			bitboard.NewMove(enums.E2, enums.G1, enums.Quiet),
			fen.FEN2Bitboard("8/8/8/8/1b6/2N5/8/4K1N1 w - - 0 1"),
			enums.WhiteKnight,
			"Ne2",
		},
		{
			bitboard.NewMove(enums.B7, enums.A6, enums.Capture),
			fen.FEN2Bitboard("2k5/Qr6/Q7/8/8/8/8/3R4 w - - 0 1"),
			enums.WhiteQueen,
			"Q6xb7",
		},
		{
			bitboard.NewMove(enums.E8, enums.D7, enums.QueenPromoCapture),
			fen.FEN2Bitboard("4b3/3P1P2/8/8/8/8/8/8 w - - 0 1"),
			enums.WhitePawn,
			"dxe8=Q",
		},
		{
			bitboard.NewMove(enums.E4, enums.F6, enums.Capture),
			fen.FEN2Bitboard("rnbqkb1r/pppppppp/5n2/8/3PP3/8/PPP2PPP/RNBQKBNR b KQkq - 0 1"),
			enums.BlackKnight,
			"Nxe4",
		},
		{
			bitboard.NewMove(enums.E5, enums.D4, enums.Capture),
			fen.FEN2Bitboard("8/8/8/4p3/3P4/8/8/8 w - - 0 1"),
			enums.WhitePawn,
			"dxe5",
		},
	}
	for _, tc := range testcases {
		tc.bb.GenLegalMoves()
		got := Move2SAN(tc.move, tc.bb.Pieces, tc.bb.LegalMoves, tc.pt)
		if got != tc.expected {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
	}
}
