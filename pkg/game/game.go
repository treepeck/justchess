package game

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
)

type Game struct {
	Result   enums.Result
	Bitboard *bitboard.Bitboard
	// Slice of the completed moves.
	Moves []bitboard.Move
}

func NewGame(r enums.Result, bb *bitboard.Bitboard) *Game {
	return &Game{
		Result:   r,
		Bitboard: bb,
	}
}

func (g *Game) ProcessMove(m bitboard.Move) {
	for _, legalMove := range g.Bitboard.LegalMoves {
		// Check if the move is legal.
		if m.To() != legalMove.To() || m.From() != legalMove.From() {
			continue
		}
		// The default move type for a promotion is QueenPromo, but the player might
		// want to promote to the other piece.
		if m.Type() >= enums.KnightPromo && m.Type() <= enums.QueenPromo &&
			legalMove.Type() == enums.QueenPromo {
			legalMove = m
			// Same for capture promotion.
		} else if m.Type() >= enums.KnightPromoCapture && m.Type() <=
			enums.QueenPromoCapture && legalMove.Type() == enums.QueenPromoCapture {
			legalMove = m
		}
		g.Bitboard.MakeMove(legalMove)
		c, opC := g.Bitboard.ActiveColor, g.Bitboard.ActiveColor^1
		pt := g.Bitboard.GetPieceTypeFromSquare(m.To())
		// TRICK: Store the current castling rights.
		// The checked king will not be able to castle on the next move,
		// but should be able to castle later if the king and rooks did not make moves.
		crCopy := g.Bitboard.CastlingRights
		g.Bitboard.ActiveColor = opC
		// To determine if the last m was a check, generate possible moves
		// for the moved piece.
		occupied := g.Bitboard.Pieces[0] | g.Bitboard.Pieces[1] | g.Bitboard.Pieces[2] |
			g.Bitboard.Pieces[3] | g.Bitboard.Pieces[4] | g.Bitboard.Pieces[5] |
			g.Bitboard.Pieces[6] | g.Bitboard.Pieces[7] | g.Bitboard.Pieces[8] |
			g.Bitboard.Pieces[9] | g.Bitboard.Pieces[10] | g.Bitboard.Pieces[11]
		isCheck := bitboard.GenAttackedSquares(1<<m.To(), occupied, pt)&
			g.Bitboard.Pieces[10+opC] != 0
		if isCheck {
			g.Bitboard.CastlingRights[0+c] = false
			g.Bitboard.CastlingRights[2+c] = false
		}
		// Generate legal moves for the next color.
		g.Bitboard.GenLegalMoves()
		g.Bitboard.CastlingRights = crCopy
		if len(g.Bitboard.LegalMoves) == 0 {
			if isCheck {
				g.Result = enums.Checkmate
			} else {
				g.Result = enums.Stalemate
			}
		}
	}
}
