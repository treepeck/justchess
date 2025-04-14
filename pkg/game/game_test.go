package game

import (
	"justchess/pkg/chess"
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
	"testing"
)

func TestCompressMoves(t *testing.T) {
	testcases := []struct {
		name     string
		move     chess.CompletedMove
		expected int
	}{
		{
			"fxg3",
			chess.CompletedMove{
				Move:     bitboard.NewMove(enums.G3, enums.F2, enums.Capture),
				SAN:      "fxg3",
				FEN:      "3rk2r/pp1b1pp1/3Qp2p/2PnN3/2P5/6P1/Pq4PP/R4R1K b - - 0 1",
				TimeLeft: 610,
			},
			0x2624356,
		},
		{
			"cxd8=Q#",
			chess.CompletedMove{
				Move:     bitboard.NewMove(enums.D8, enums.C7, enums.QueenPromoCapture),
				SAN:      "cxd8=Q#",
				FEN:      "3Pk2r/p2b1pp1/3Qp2p/2PnN3/8/6P1/Pq4PP/R4R1K b - - 0 1",
				TimeLeft: 10,
			},
			0xADCBB,
		},
	}

	for _, tc := range testcases {
		got := compressMoves([]chess.CompletedMove{tc.move})[0]

		if got != tc.expected {
			t.Fatalf("expected: %d, got: %d\n", tc.expected, got)
		}
	}
}

func BenchmarkCompressMoves(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compressMoves([]chess.CompletedMove{{
			Move:     bitboard.NewMove(enums.G3, enums.F2, enums.Capture),
			SAN:      "fxg3",
			FEN:      "3rk2r/pp1b1pp1/3Qp2p/2PnN3/2P5/6P1/Pq4PP/R4R1K b - - 0 1",
			TimeLeft: 610,
		}})
	}
}
