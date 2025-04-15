package bitboard

import (
	"justchess/pkg/chess/enums"
)

// Bitboard stores the chessboard state which can be converted to a FEN.
type Bitboard struct {
	Pieces      [12]uint64
	ActiveColor enums.Color
	// [0] - White king castle.
	// [1] - Black king castle.
	// [2] - White queen castle.
	// [3] - Black queen castle.
	CastlingRights [4]bool
	// The index of a square which a pawn just passed while performing a
	// double push forward. If there isn't such square, will be equal to enums.NoSquare (-1).
	EPTarget int
	// Number of half moves since the last capture or pawn move.
	// Used to implement the fifty-move rule.
	HalfmoveCnt int
	// Number of completed full moves.
	FullmoveCnt int
	// Legal moves for the curently active color.
	LegalMoves []Move
}

var (
	promoPieces = [4]enums.PieceType{enums.WhiteKnight, enums.WhiteBishop,
		enums.WhiteRook, enums.WhiteQueen}
)

func NewBitboard(pieces [12]uint64, ac enums.Color,
	cr [4]bool, ep, half, full int) *Bitboard {
	return &Bitboard{
		Pieces:         pieces,
		ActiveColor:    ac,
		CastlingRights: cr,
		EPTarget:       ep,
		HalfmoveCnt:    half,
		FullmoveCnt:    full,
		LegalMoves:     make([]Move, 0),
	}
}

// GenLegalMoves generates legal moves for the current active color.
// The legal moves generation takes 3 steps:
//
//  1. Generate pseudo-legal moves on a board for all pieces except king;
//  2. Filter down illegal moves;
//  3. Generate legal moves for a king. Since generation of the king's pseudo-legal moves
//     is anyway a demanding operation, it is usefull to generate legal moves straightaway.
func (bb *Bitboard) GenLegalMoves() {
	pseudoLegal := make([]Move, 0)
	c, opC := bb.ActiveColor, bb.ActiveColor^1

	// Make a bitboard with all allied pieces.
	allies := bb.Pieces[0+c] | bb.Pieces[2+c] | bb.Pieces[4+c] | bb.Pieces[6+c] |
		bb.Pieces[8+c] | bb.Pieces[10+c]
	// Make a bitboard with all enemy pieces.
	enemies := bb.Pieces[0+opC] | bb.Pieces[2+opC] | bb.Pieces[4+opC] | bb.Pieces[6+opC] |
		bb.Pieces[8+opC] | bb.Pieces[10+opC]
	// Generate pseudo legal moves for pawns.
	pseudoLegal = append(pseudoLegal, genPawnsPseudoLegalMoves(bb.Pieces[0+c], allies, enemies, c,
		bb.EPTarget)...)
	// Take each piece type except the king and pawns.
	for i := 2; i < len(bb.Pieces)-2; i++ {
		if i%2 != int(bb.ActiveColor) || bb.Pieces[i] == 0 {
			continue
		}
		pseudoLegal = append(pseudoLegal, genPseudoLegalMoves(enums.PieceType(i),
			bb.Pieces[i], allies, enemies)...)
	}

	bb.LegalMoves = bb.filterIllegalMoves(pseudoLegal)

	// Exclude the king from occupied, otherwise the king will be able
	// to move on attacked squares behind himself.
	king := bb.Pieces[10+c]
	bb.Pieces[10+c] ^= king
	attacked := GenAttackedSquares(bb.Pieces, opC)
	// Restore king position.
	bb.Pieces[10+c] ^= king
	bb.LegalMoves = append(bb.LegalMoves, genKingLegalMoves(bb.Pieces[10+c], allies,
		enemies, attacked, bb.CastlingRights[c], bb.CastlingRights[c+2], c)...)
}

// MakeMove performs the move affecting only the pieces' placement.
func (bb *Bitboard) MakeMove(m Move) {
	var from, to uint64 = 1 << m.From(), 1 << m.To()
	fromTo := from ^ to
	movedPT := GetPieceOnSquare(from, bb.Pieces)

	switch m.Type() {
	case enums.Capture:
		bb.Pieces[GetPieceOnSquare(to, bb.Pieces)] ^= to

	case enums.EPCapture:
		if movedPT == enums.WhitePawn {
			bb.Pieces[enums.BlackPawn] ^= to >> 8
		} else {
			bb.Pieces[enums.WhitePawn] ^= to << 8
		}

	case enums.KingCastle:
		rookFrom, rookTo := to<<1, to>>1
		bb.Pieces[movedPT-4] ^= rookFrom ^ rookTo

	case enums.QueenCastle:
		rookFrom, rookTo := to>>2, to<<1
		bb.Pieces[movedPT-4] ^= rookFrom ^ rookTo

	case enums.KnightPromo, enums.BishopPromo,
		enums.RookPromo, enums.QueenPromo:
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+(promoPieces[m.Type()-6])] ^= to

	case enums.KnightPromoCapture, enums.BishopPromoCapture,
		enums.RookPromoCapture, enums.QueenPromoCapture:
		bb.Pieces[GetPieceOnSquare(to, bb.Pieces)] ^= to
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+(promoPieces[m.Type()-10])] ^= to
	}

	bb.Pieces[movedPT] ^= fromTo
}

// SetCastlingRights sets the castling rights to false if the king has moved,
// or the rooks are not on their standart positions.
// pt stores the type of the last moved piece.
func (bb *Bitboard) SetCastlingRights(pt enums.PieceType) {
	if pt == enums.WhiteKing || pt == enums.BlackKing {
		bb.CastlingRights[0+bb.ActiveColor] = false
		bb.CastlingRights[2+bb.ActiveColor] = false
	}
	if bb.Pieces[enums.WhiteRook]&wkr == 0 {
		bb.CastlingRights[2] = false
	}
	if bb.Pieces[enums.WhiteRook]&wqr == 0 {
		bb.CastlingRights[0] = false
	}
	if bb.Pieces[enums.BlackRook]&bkr == 0 {
		bb.CastlingRights[3] = false
	}
	if bb.Pieces[enums.BlackRook]&bqr == 0 {
		bb.CastlingRights[1] = false
	}
}

// SetEPTarget sets the en passant target square after completing the move.
func (bb *Bitboard) SetEPTarget(lastMove Move) {
	// Reset the en passant target since the en passant capture is possible only
	// for 1 move.
	bb.EPTarget = enums.NoSquare

	// After double pawn push, set the en passant target.
	if lastMove.Type() == enums.DoublePawnPush {
		if bb.ActiveColor == enums.White {
			bb.EPTarget = lastMove.To() - 8
		} else {
			bb.EPTarget = lastMove.To() + 8
		}
	}
}

// SetHalfmoveCnt sets the halfmove counter depending on the last move's type.
func (bb *Bitboard) SetHalfmoveCnt(pt enums.PieceType, mt enums.MoveType) {
	if mt >= enums.Capture {
		bb.HalfmoveCnt = 0
	} else if pt != enums.WhitePawn && pt != enums.BlackPawn {
		bb.HalfmoveCnt++
	}
}

// IsMoveLegal returns true if the bb stores the simmilar move in LegalMoves field.
func (bb *Bitboard) IsMoveLegal(m Move) bool {
	for _, lm := range bb.LegalMoves {
		if m.To() != lm.To() || m.From() != lm.From() {
			continue
		}

		if m.Type() < enums.KnightPromo {
			return m.Type() == lm.Type()
		}

		// FIXME: when moves are generated, the move type for a promotion is either QueenPromo
		// or QueenPromoCapture is case of capture promotion. But the player might want
		// to promote to the other piece.
		isValidPromo := m.Type() >= enums.KnightPromo && m.Type() <= enums.QueenPromo &&
			lm.Type() == enums.QueenPromo

		isValidPromoCapture := m.Type() >= enums.KnightPromoCapture &&
			m.Type() <= enums.QueenPromoCapture && lm.Type() == enums.QueenPromoCapture

		return isValidPromo || isValidPromoCapture
	}
	return false
}

// GetPieceOnSquare returns the type of the piece that stands on the specified square.
// If there is no piece on the square, returns WhitePawn.
func GetPieceOnSquare(square uint64, pieces [12]uint64) enums.PieceType {
	for pt, bitboard := range pieces {
		if square&bitboard != 0 {
			return enums.PieceType(pt)
		}
	}
	return enums.WhitePawn
}

func (bb *Bitboard) CalculateMaterial() (mat int) {
	material := map[enums.PieceType]int{
		enums.WhitePawn:   1,
		enums.BlackPawn:   1,
		enums.WhiteKnight: 3,
		enums.BlackKnight: 3,
		enums.WhiteBishop: 3,
		enums.BlackBishop: 3,
		enums.WhiteRook:   5,
		enums.BlackRook:   5,
		enums.WhiteQueen:  9,
		enums.BlackQueen:  9,
	}
	for pt := 0; pt < 10; pt++ {
		for bitboard := bb.Pieces[pt]; bitboard != 0; bitboard &= bitboard - 1 {
			mat += material[enums.PieceType(pt)]
		}
	}
	return
}

// filterIllegalMoves sequentially performs the pseudo-legal moves
// on a board copy and rejects the moves which lead to the checked king.
func (bb *Bitboard) filterIllegalMoves(pseudoLegal []Move) (legal []Move) {
	boardCopy := bb.Pieces
	for _, move := range pseudoLegal {
		bb.MakeMove(move)
		// If the allied king is not in check, the move is legal.
		if GenAttackedSquares(bb.Pieces, bb.ActiveColor^1)&
			bb.Pieces[10+bb.ActiveColor] == 0 {
			legal = append(legal, move)
		}
		// Restore the board.
		bb.Pieces = boardCopy
	}
	return
}
