package bitboard

import (
	"justchess/pkg/game/enums"
	"justchess/pkg/game/helpers"
)

// Bitboard stores the chessboard state which is described by FEN.
type Bitboard struct {
	// Piece placement. First index - color, second index - piece type.
	Pieces      [2][7]uint64
	ActiveColor enums.Color
	// [0] - White king castle.
	// [1] - White queen castle.
	// [2] - Black king castle.
	// [3] - Black queen castle.
	CastlingRights [4]bool
	// The index of a square which a pawn just passed while performing a
	// double push forward. If there isn`t such square, will be equal to -1.
	EPTarget int
	// Number of half moves since the last capture or pawn move.
	// It is used to implement the fifty-move rule.
	HalfmoveCnt int
	// Number of completed full moves.
	FullmoveCnt int
}

func NewBitboard(pieces [2][7]uint64, ac enums.Color,
	cr [4]bool, ep, half, full int) *Bitboard {
	b := &Bitboard{
		Pieces:         pieces,
		ActiveColor:    ac,
		CastlingRights: cr,
		EPTarget:       ep,
		HalfmoveCnt:    half,
		FullmoveCnt:    full,
	}
	return b
}

// MakeMove performs the move on a COPY of the current board.
// Returns the modified board copy.
// Used to check if the move is legal.
func (bb *Bitboard) MakeMove(move helpers.Move) (boardCopy [2][7]uint64) {
	copy(boardCopy[:], bb.Pieces[:])
	from := uint64(1) << move.From()
	to := uint64(1) << move.To()
	fromTo := from ^ to
	pt := helpers.GetPieceTypeFromSquare(bb.Pieces, move.From())
	boardCopy[bb.ActiveColor][pt] ^= fromTo
	boardCopy[bb.ActiveColor][0] ^= fromTo
	return
}

func (bb *Bitboard) GenLegalMoves() []helpers.Move {
	c := bb.ActiveColor
	opC := c.Inverse()
	// First of all the pseudo legal moves must be generated.
	pl := make([]helpers.Move, 0)
	for i := 1; i < 6; i++ { // Take each piece type except the king.
		pieces := helpers.GetIndicesFromBitboard(bb.Pieces[c][i])
		for _, piece := range pieces {
			allies := bb.Pieces[c][0]    // All pieces of the active color.
			enemies := bb.Pieces[opC][0] // All enemy pieces.
			pl = append(pl, genPseudoLegalMoves(i, c, piece, allies, enemies)...)
		}
	}
	l := bb.filterIllegalMoves(pl)
	// Add legal moves for king.
	// To prevent the king from moving into attacked squares, exclude the king from
	// occupied, otherwise the king will be able to move behind himself on attacked squares.
	occupied := bb.Pieces[c][1] | bb.Pieces[c][2] | bb.Pieces[c][3] |
		bb.Pieces[c][4] | bb.Pieces[c][5] | bb.Pieces[opC][0]
	attacked := genAttackedSquares(opC, bb.Pieces[opC][1:], occupied)
	can00, can000 := bb.CastlingRights[c], bb.CastlingRights[c+2]
	kingPos := helpers.GetIndicesFromBitboard(bb.Pieces[c][6])[0]
	l = append(l, genKingLegalMoves(kingPos, bb.Pieces[c][0],
		bb.Pieces[opC][0], attacked, can00, can000)...)
	return l
}

func (bb *Bitboard) filterIllegalMoves(pl []helpers.Move) (l []helpers.Move) {
	opColor := bb.ActiveColor.Inverse()
	for _, move := range pl {
		board := bb.MakeMove(move)
		enemies := board[opColor][1:]         // All enemy pieces.
		occupied := board[0][0] | board[1][0] // All occupied squares.
		// If the allied king is not in check, the move is legal.
		if genAttackedSquares(opColor, enemies, occupied)&
			bb.Pieces[bb.ActiveColor][6] == 0 {
			l = append(l, move)
		}
	}
	return
}
