package bitboard

import "justchess/pkg/game/enums"

// Bitboard stores the chessboard state which is described by FEN.
type Bitboard struct {
	// Piece placement.
	Pieces      [12]uint64
	ActiveColor enums.Color
	// [0] - White king castle.
	// [1] - Black king castle.
	// [2] - White queen castle.
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
	LegalMoves  []Move
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
		LegalMoves:     make([]Move, 0),
	}
	b.GenLegalMoves()
	return b
}

// MakeMove changes the current board state by performing the move.
func (bb *Bitboard) MakeMove(m Move, pt enums.PieceType) {
	var from, to uint64 = 1 << m.From(), 1 << m.To()
	fromTo := from ^ to
	switch m.Type() {
	case enums.Quiet, enums.DoublePawnPush:
		bb.Pieces[pt] ^= fromTo
	case enums.KingCastle:
		bb.Pieces[pt] ^= fromTo
		// When making a O-O move, the king will always be 1 square left to the rook.
		rookFrom, rookTo := to<<1, to>>1
		bb.Pieces[6+bb.ActiveColor] ^= rookFrom ^ rookTo
	case enums.QueenCastle:
		bb.Pieces[pt] ^= fromTo
		// When making a O-O-O move, the king will always be 2 squares right to the rook.
		rookFrom, rookTo := to>>2, to<<1
		bb.Pieces[6+bb.ActiveColor] ^= rookFrom ^ rookTo
	case enums.Capture:
		bb.Pieces[pt] ^= fromTo
		bb.Pieces[pt] ^= to
	case enums.EPCapture:
		if bb.ActiveColor == enums.White {
			bb.Pieces[0+(bb.ActiveColor^1)] ^= to >> 8
		} else {
			bb.Pieces[0+(bb.ActiveColor^1)] ^= to << 8
		}
		bb.Pieces[0+bb.ActiveColor] ^= fromTo
	case enums.KnightPromo:
		bb.Pieces[pt] ^= from
		bb.Pieces[2+bb.ActiveColor] ^= to
	case enums.BishopPromo:
		bb.Pieces[pt] ^= from
		bb.Pieces[4+bb.ActiveColor] ^= to
	case enums.RookPromo:
		bb.Pieces[pt] ^= from
		bb.Pieces[6+bb.ActiveColor] ^= to
	case enums.QueenPromo:
		bb.Pieces[pt] ^= from
		bb.Pieces[8+bb.ActiveColor] ^= to
	}
}

// GenLegalMoves generates legal moves for the current active color.
func (bb *Bitboard) GenLegalMoves() {
	// First of all the pseudo legal moves must be generated.
	pseudoLegal := make([]Move, 0)
	var c, opC enums.Color = bb.ActiveColor, bb.ActiveColor ^ 1
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
		for bitboard := bb.Pieces[i]; bitboard != 0; bitboard &= bitboard - 1 {
			from := GetLSB(bitboard)
			pseudoLegal = append(pseudoLegal, genPseudoLegalMoves(enums.PieceType(i),
				from, allies, enemies)...)
		}
	}
	bb.LegalMoves = bb.filterIllegalMoves(pseudoLegal)
	// Add legal moves for king.
	// Exclude the king from occupied, otherwise the king will be able
	// to move on attacked squares behind himself.
	attacked := genAttackedSquaresBySide([6]uint64{
		bb.Pieces[0+opC], bb.Pieces[2+opC], bb.Pieces[4+opC],
		bb.Pieces[6+opC], bb.Pieces[8+opC], bb.Pieces[10+opC],
	}, occupied^bb.Pieces[10+c], opC)
	kingPos := GetLSB(bb.Pieces[10+c])
	bb.LegalMoves = append(bb.LegalMoves, genKingLegalMoves(kingPos, allies,
		enemies, attacked, bb.CastlingRights[c], bb.CastlingRights[c+2])...)
}

func (bb *Bitboard) filterIllegalMoves(pseudoLegal []Move) (legal []Move) {
	boardCopy := bb.Pieces
	var c, opC enums.Color = bb.ActiveColor, bb.ActiveColor ^ 1
	for _, move := range pseudoLegal {
		pt := GetPieceTypeFromSquare(move.From(), bb.Pieces)
		bb.MakeMove(move, pt)
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
		// Restore the board.
		bb.Pieces = boardCopy
	}
	return
}

// GetPieceTypeFromSquare returns the type of the piece that stands on the specified square.
// If there is no piece on the square, returns WhitePawn.
func GetPieceTypeFromSquare(square int, pieces [12]uint64) enums.PieceType {
	var sqBB uint64 = 1 << square
	for pt, bitboard := range pieces {
		if sqBB&bitboard != 0 {
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
