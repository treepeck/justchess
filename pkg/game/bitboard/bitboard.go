package bitboard

import "justchess/pkg/game/enums"

// Bitboard stores the chessboard state which is described by FEN.
type Bitboard struct {
	// Piece placement.
	Pieces      [12]uint64
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

func NewBitboard(pieces [12]uint64, ac enums.Color,
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

// MakeMove performs the move on the current board state.
func (bb *Bitboard) MakeMove(move Move) {
	var from uint64 = 1 << move.From()
	var to uint64 = 1 << move.To()
	fromTo := from ^ to
	bb.Pieces[bb.getPieceTypeFromSquare(move.From())] ^= fromTo
	if move.MoveType() == enums.Capture {
		bb.Pieces[bb.getPieceTypeFromSquare(move.To())] ^= to
	}
}

func (bb *Bitboard) GenLegalMoves() []Move {
	// First of all the pseudo legal moves must be generated.
	pseudoLegal := make([]Move, 0)
	c := bb.ActiveColor
	opC := c ^ 1
	var allies, enemies, occupied uint64
	allies |= bb.Pieces[0+c] | bb.Pieces[2+c] | bb.Pieces[4+c] |
		bb.Pieces[6+c] | bb.Pieces[8+c] | bb.Pieces[10+c]
	enemies |= bb.Pieces[0+opC] | bb.Pieces[2+opC] | bb.Pieces[4+opC] |
		bb.Pieces[6+opC] | bb.Pieces[8+opC] | bb.Pieces[10+opC]
	occupied = allies | enemies
	// Take each piece type except the king.
	for i := 0; i < len(bb.Pieces)-2; i++ {
		if i%2 != int(c) {
			continue
		}
		bitboard := bb.Pieces[i]
		for from := GetLSB(bitboard); bitboard != 0; from = GetLSB(bitboard) {
			pseudoLegal = append(pseudoLegal, genPseudoLegalMoves(enums.PieceType(i),
				from, allies, enemies)...)
			bitboard &= bitboard - 1
		}
	}
	legal := bb.filterIllegalMoves(pseudoLegal)
	// Add legal moves for king.
	// To prevent the king from moving into attacked squares, exclude the king from
	// occupied, otherwise the king will be able to move behind himself on attacked squares.
	attacked := genAttackedSquaresBySide([6]uint64{
		bb.Pieces[0+opC], bb.Pieces[2+opC], bb.Pieces[4+opC],
		bb.Pieces[6+opC], bb.Pieces[8+opC], bb.Pieces[10+opC],
	}, occupied^bb.Pieces[10+c], opC)
	kingPos := GetLSB(bb.Pieces[10+c])
	legal = append(legal, genKingLegalMoves(kingPos, allies, enemies,
		attacked, bb.CastlingRights[c], bb.CastlingRights[c+2])...)
	return legal
}

func (bb *Bitboard) filterIllegalMoves(pseudoLegal []Move) (legal []Move) {
	boardCopy := bb.Pieces
	c := bb.ActiveColor
	opC := c ^ 1
	for _, move := range pseudoLegal {
		bb.MakeMove(move)
		occupied := bb.Pieces[0] | bb.Pieces[1] | bb.Pieces[2] | bb.Pieces[3] |
			bb.Pieces[4] | bb.Pieces[5] | bb.Pieces[6] | bb.Pieces[7] |
			bb.Pieces[8] | bb.Pieces[9] | bb.Pieces[10] | bb.Pieces[11]
		// Get all attacked squares on a new position.
		attacked := genAttackedSquaresBySide([6]uint64{
			bb.Pieces[0+opC], bb.Pieces[2+opC], bb.Pieces[4+opC],
			bb.Pieces[6+opC], bb.Pieces[8+opC], bb.Pieces[10+opC],
		}, occupied, opC)
		// If the allied king is not in check, the move is legal.
		if attacked&bb.Pieces[10+c] == 0 {
			legal = append(legal, move)
		}
		// Restore board state.
		bb.Pieces = boardCopy
	}
	return
}

// getPieceTypeFromSquare returns the type of the piece that stands on the specified square.
// If there is no piece on the square, returns WhitePawn.
func (bb *Bitboard) getPieceTypeFromSquare(square int) enums.PieceType {
	for pt, bitboard := range bb.Pieces {
		for i := GetLSB(bitboard); bitboard != 0; i = GetLSB(bitboard) {
			if i == square {
				return enums.PieceType(pt)
			}
			bitboard &= bitboard - 1
		}
	}
	return enums.WhitePawn
}
