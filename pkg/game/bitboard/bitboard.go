package bitboard

import (
	"justchess/pkg/game/enums"
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
	// Generated automatically after each move.
	LegalMoves []Move
}

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

// MakeMove performs the move on a bitboard. Only affects piece placement.
func (bb *Bitboard) MakeMove(m Move) {
	var from, to uint64 = 1 << m.From(), 1 << m.To()
	fromTo := from ^ to
	movedPT := GetPieceTypeFromSquare(from, bb.Pieces)

	switch m.Type() {
	case enums.Capture:
		bb.Pieces[GetPieceTypeFromSquare(to, bb.Pieces)] ^= to

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

	case enums.KnightPromo:
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+enums.WhiteKnight] ^= to

	case enums.BishopPromo:
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+enums.WhiteBishop] ^= to

	case enums.RookPromo:
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+enums.WhiteRook] ^= to

	case enums.QueenPromo:
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+enums.WhiteQueen] ^= to

	case enums.KnightPromoCapture:
		bb.Pieces[GetPieceTypeFromSquare(to, bb.Pieces)] ^= to
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+enums.WhiteKnight] ^= to

	case enums.BishopPromoCapture:
		bb.Pieces[GetPieceTypeFromSquare(to, bb.Pieces)] ^= to
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+enums.WhiteBishop] ^= to

	case enums.RookPromoCapture:
		bb.Pieces[GetPieceTypeFromSquare(to, bb.Pieces)] ^= to
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+enums.WhiteRook] ^= to

	case enums.QueenPromoCapture:
		bb.Pieces[GetPieceTypeFromSquare(to, bb.Pieces)] ^= to
		bb.Pieces[movedPT] ^= to
		bb.Pieces[movedPT+enums.WhiteQueen] ^= to
	}

	bb.Pieces[movedPT] ^= fromTo
}

// GenLegalMoves generates legal moves for the current active color.
func (bb *Bitboard) GenLegalMoves() {
	// First of all generate the pseudo-legal moves.
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

	// Secondly, reject illegal moves.
	bb.LegalMoves = bb.filterIllegalMoves(pseudoLegal)

	// Finally add legal moves for the king.
	// Exclude the king from occupied, otherwise the king will be able
	// to move on attacked squares behind himself.
	var king = bb.Pieces[10+c]
	bb.Pieces[10+c] ^= king
	attacked := GenAttackedSquares(bb.Pieces, opC)
	// Restore king position.
	bb.Pieces[10+c] ^= king
	bb.LegalMoves = append(bb.LegalMoves, genKingLegalMoves(bb.Pieces[10+c], allies,
		enemies, attacked, bb.CastlingRights[c], bb.CastlingRights[c+2], c)...)
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

// GetPieceTypeFromSquare returns the type of the piece that stands on the specified square.
// If there is no piece on the square, returns WhitePawn.
func GetPieceTypeFromSquare(square uint64, pieces [12]uint64) enums.PieceType {
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
