package bitboard

import (
	"justchess/pkg/game/enums"
	"justchess/pkg/game/helpers"
)

// Bitboard stores only fields that can be serialized to a Forsyth-Edwards Notation.
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

// MakeMove makes the move on a bitboard copy.
func MakeMove(pieces []uint64, move helpers.Move) {
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

// filterIllegalMoves filtes the moves by finding all attacked by enemies
// squares after making the move. If the allies king is attacked, move is not legal.
func (bb *Bitboard) filterIllegalMoves(pseudoLegal []helpers.Move,
) (legal []helpers.Move) {
	before := bb.Pieces[:]
	var allies [6]uint64
	allies[0] = bb.Pieces[2+bb.ActiveColor]  // Allied pawns.
	allies[1] = bb.Pieces[4+bb.ActiveColor]  // Allied knights.
	allies[2] = bb.Pieces[6+bb.ActiveColor]  // Allied bishops.
	allies[3] = bb.Pieces[8+bb.ActiveColor]  // Allied rooks.
	allies[4] = bb.Pieces[10+bb.ActiveColor] // Allied queens.
	allies[5] = bb.Pieces[12+bb.ActiveColor] // Allied king.
	for _, move := range pseudoLegal {
		MakeMove(before, move)
		isChecked := true
		if genAttackedSquares(bb.ActiveColor.Inverse(), allies,
			bb.Pieces[0]|bb.Pieces[1])&allies[5] != 0 {
			isChecked = false
		}
		// Restore the board state.
		copy(bb.Pieces[:], before)
		if !isChecked {
			legal = append(legal, move)
		}
	}
	return
}

// genLegalMove generates pseudo-legal moves for all pieces of the active color.
// After that the generated moves are filtered down to legal moves by
// filterIllegalMoves function.
func (bb *Bitboard) genLegalMoves() {
	psm := make(map[int][]helpers.Move)
	for i := 2; i < 14; i++ {
		if (bb.ActiveColor == enums.Black && i%2 == 0) ||
			(bb.ActiveColor == enums.White && i%2 != 0) {
			continue
		}
		for _, piece := range helpers.GetIndicesFromBitboard(bb.Pieces[i]) {
			psm[piece] = genPseudoLegalMoves(enums.PieceType(i), piece, bb.Pieces[0],
				bb.Pieces[1])
		}
	}
}
