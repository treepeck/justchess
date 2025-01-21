package bitboard

import (
	"justchess/pkg/game/enums"
	"justchess/pkg/game/helpers"
)

// Bitboard represents a chessboard in a bitboard manner.
type Bitboard struct {
	// Pieces stores a piece placement bitboard for each piece type and for all pieces of both colors.
	// Bitboards can be accessed by indexes:
	// [0] - All white pieces;
	// [1] - All black pieces;
	// [2] - White pawns;
	// [3] - Black pawns;
	// [4] - White Knights;
	// [5] - Black Knights;
	// [6] - White Bishops;
	// [7] - Black Bishops;
	// [8] - White Rooks;
	// [9] - Black Rooks;
	// [10] - White Queens;
	// [11] - Black Queens;
	// [12] - White King;
	// [13] - Black King.
	// In each board LSBs represent white pieces, MSBs - black pieces.
	Pieces      [14]uint64
	ActiveColor enums.Color
	// CastlingRights stores the castling rights:
	// [0] - White king castle. (0-0)
	// [1] - White queen castle. (0-0-0)
	// [2] - Black king castle. (0-0)
	// [3] - Black queen castle. (0-0-0)
	CastlingRights [4]bool
	// EpTarget stores the index of a square which a pawn just passed while performing a
	// double push forward.
	// If there isn`t EP target, the EpTarget will be equal to -1.
	EpTarget int
	// HalfmoveClk stores the number of half moves since the last capture or pawn move.
	// It is used to implement the fifty-move rule.
	HalfmoveClk int
	// FullmoveClk stores the number of full moves.
	FullmoveClk int
}

// NewBitboard returns initialized Bitboard.
func NewBitboard(pieces [14]uint64, activeColor enums.Color,
	castlingRights [4]bool, epTarget, hclk, fclk int) *Bitboard {
	b := &Bitboard{
		Pieces:         pieces,
		CastlingRights: castlingRights,
		EpTarget:       epTarget,
		HalfmoveClk:    hclk,
		FullmoveClk:    fclk,
	}
	return b
}

// MakeMove makes the move on a bitboard COPY.
func MakeMove(pieces [14]uint64, move helpers.Move) {
	from := uint64(1) << move.From
	to := uint64(1) << move.To
	fromTo := from ^ to
	pieces[move.PieceType] ^= fromTo
	pieces[move.Color] ^= fromTo

	switch move.MoveType {
	case enums.Capture, enums.EpCapture:
		pieces[move.CapturedPieceType] ^= to
		pieces[move.Color.Inverse()] ^= to
	}
}

// genWhitePseudoLegalMoves generates pseudo-legal moves for all white pieces.
// Generated moves are filtered down to legal moves by filterMoves function.
func (bb *Bitboard) genWhitePseudoLegalMoves() map[int][]helpers.Move {
	psm := make(map[int][]helpers.Move)
	for i := 2; i < 14; i += 2 {
		for _, piece := range helpers.GetIndicesFromBitboard(bb.Pieces[i]) {
			psm[piece] = genPseudoLegalMoves(enums.PieceType(i), piece, bb.Pieces[0],
				bb.Pieces[1])
		}
	}
	return psm
}

// filterMoves filters down the pseudo-legal moves to legal only.
// func (bb *Bitboard) filterMoves() {
// 	psm := bb.genWhitePseudoLegalMoves()
// 	piecesCopy := bb.Copy()
// 	for from, moves := range psm {
// 		for _, move := range moves {
// 			MakeMove(piecesCopy, move)
// 		}
// 	}
// }

// TODO: initBoards might be deleted.
// initBoards creates 14 bitboards with standart piece positions.
// func initBoards() [14]uint64 {
// 	var pieces [14]uint64
// 	pieces[0] = 0xFFFF
// 	pieces[1] = 0xFFFF000000000000
// 	pieces[2] = 0xFF
// 	pieces[3] = 0xFF00000000000000
// 	pieces[4] = 0x42
// 	pieces[5] = 0x4200000000000000
// 	pieces[6] = 0x24
// 	pieces[7] = 0x2400000000000000
// 	pieces[8] = 0x81
// 	pieces[9] = 0x8100000000000000
// 	pieces[10] = 0x10
// 	pieces[11] = 0x1000000000000000
// 	pieces[12] = 0x8
// 	pieces[13] = 0x08000000000000000
// 	return pieces
// }
