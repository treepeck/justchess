package san

import (
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
	"justchess/pkg/chess/fen"
	"testing"
)

func TestMove2SAN(t *testing.T) {
	testcases := []struct {
		move        bitboard.Move
		bb          *bitboard.Bitboard
		pt          enums.PieceType
		isCheck     bool
		isCheckmate bool
		expected    string
	}{
		{
			bitboard.NewMove(enums.E2, enums.C3, enums.Quiet),
			fen.FEN2Bitboard("8/8/8/8/8/2N5/8/4K1N1 w - - 0 1"),
			enums.WhiteKnight,
			false,
			false,
			"Nce2",
		},
		{
			// Similar case to the previous one, except the knight c3 is pinned by the black
			// bishop, so the disambiguation is not needed.
			bitboard.NewMove(enums.E2, enums.G1, enums.Quiet),
			fen.FEN2Bitboard("8/8/8/8/1b6/2N5/8/4K1N1 w - - 0 1"),
			enums.WhiteKnight,
			false,
			false,
			"Ne2",
		},
		{
			bitboard.NewMove(enums.B7, enums.A6, enums.Capture),
			fen.FEN2Bitboard("2k5/Qr6/Q7/8/8/8/8/3R4 w - - 0 1"),
			enums.WhiteQueen,
			true,
			true,
			"Q6xb7#",
		},
		{
			bitboard.NewMove(enums.E8, enums.D7, enums.QueenPromoCapture),
			fen.FEN2Bitboard("4b3/3P1P2/8/8/8/8/8/8 w - - 0 1"),
			enums.WhitePawn,
			false,
			false,
			"dxe8=Q",
		},
		{
			bitboard.NewMove(enums.E4, enums.F6, enums.Capture),
			fen.FEN2Bitboard("rnbqkb1r/pppppppp/5n2/8/3PP3/8/PPP2PPP/RNBQKBNR b KQkq - 0 1"),
			enums.BlackKnight,
			false,
			false,
			"Nxe4",
		},
		{
			bitboard.NewMove(enums.D4, enums.E5, enums.Capture),
			fen.FEN2Bitboard("8/8/8/4p3/3P4/2K5/8/8 b - - 0 1"),
			enums.BlackPawn,
			true,
			false,
			"exd4+",
		},
		{
			bitboard.NewMove(enums.E7, enums.F7, enums.Capture),
			fen.FEN2Bitboard("r1bk3r/ppqpbQpp/2p4n/6B1/2BpP3/3P1P2/PPP3PP/RN3RK1 w - - 0 1"),
			enums.WhiteQueen,
			true,
			true,
			"Qxe7#",
		},
	}
	for _, tc := range testcases {
		tc.bb.GenLegalMoves()
		got := Move2SAN(tc.move, tc.bb.Pieces, tc.bb.LegalMoves, tc.pt, tc.isCheck, tc.isCheckmate)
		if got != tc.expected {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
	}
}
