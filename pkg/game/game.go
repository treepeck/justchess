package game

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"justchess/pkg/game/fen"
	"justchess/pkg/game/san"
)

type CompletedMove struct {
	SAN string
	// Biboard state after completing the move.
	FEN string
}

type Game struct {
	Result   enums.Result
	Bitboard *bitboard.Bitboard
	Moves    []CompletedMove
}

func NewGame(r enums.Result, bb *bitboard.Bitboard) *Game {
	return &Game{
		Result:   r,
		Bitboard: bb,
		Moves:    make([]CompletedMove, 0),
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
		ptBefore := bitboard.GetPieceTypeFromSquare(m.From(), g.Bitboard.Pieces)
		san := san.Move2SAN(legalMove, g.Bitboard.Pieces, g.Bitboard.LegalMoves, ptBefore)
		g.Bitboard.MakeMove(legalMove, ptBefore)
		c, opC := g.Bitboard.ActiveColor, g.Bitboard.ActiveColor^1
		g.Bitboard.ActiveColor = opC

		if g.isThreefoldRepetition() {
			g.Result = enums.Repetition
			break
		}

		if g.isInsufficientMaterial() {
			g.Result = enums.InsufficienMaterial
			break
		}
		// In case of promotion, the piece type will change.
		newPT := bitboard.GetPieceTypeFromSquare(m.To(), g.Bitboard.Pieces)
		// TRICK: Store the current castling rights.
		// The checked king will not be able to castle on the next move,
		// but should be able to castle later if the king and rooks did not make moves.
		crCopy := g.Bitboard.CastlingRights
		// To determine if the last m was a check, generate possible moves
		// for the moved piece.
		occupied := g.Bitboard.Pieces[0] | g.Bitboard.Pieces[1] | g.Bitboard.Pieces[2] |
			g.Bitboard.Pieces[3] | g.Bitboard.Pieces[4] | g.Bitboard.Pieces[5] |
			g.Bitboard.Pieces[6] | g.Bitboard.Pieces[7] | g.Bitboard.Pieces[8] |
			g.Bitboard.Pieces[9] | g.Bitboard.Pieces[10] | g.Bitboard.Pieces[11]

		isCheck := bitboard.GenAttackedSquares(1<<m.To(), occupied, newPT)&
			g.Bitboard.Pieces[10+opC] != 0
		if isCheck {
			g.Bitboard.CastlingRights[0+c] = false
			g.Bitboard.CastlingRights[2+c] = false
			san += "+"
		}

		// Generate legal moves for the next color.
		g.Bitboard.GenLegalMoves()
		g.Bitboard.CastlingRights = crCopy
		if len(g.Bitboard.LegalMoves) == 0 {
			if isCheck {
				g.Result = enums.Checkmate
				san = san[:len(san)-2] + "#"
			} else {
				g.Result = enums.Stalemate
			}
		}
		g.Moves = append(g.Moves, CompletedMove{
			SAN: san,
			FEN: fen.Bitboard2FEN(g.Bitboard),
		})
		break
	}
}

func (g *Game) isThreefoldRepetition() bool {
	duplicates := make(map[string]int)
	cnt := 0
	for _, move := range g.Moves {
		// The halfmove and fullmove FEN fields are omitted.
		FEN := move.FEN[0 : len(move.FEN)-4]
		if _, ok := duplicates[FEN]; !ok {
			duplicates[FEN] = 1
		} else {
			cnt++
			if cnt == 3 {
				return true
			}
		}
	}
	return false
}

// isInsufficientMaterial returns true if one of the following statements is true:
//  1. both sides have a bare king;
//  2. one side has a king and a minor piece against a bare king;
//  3. both sides have a king and a bishop, the bishops being the same color.
func (g *Game) isInsufficientMaterial() bool {
	var dark uint64 = 0xAA55AA55AA55AA55 // Mask for all dark squares.
	white, black := g.Bitboard.CalculateMaterial()
	if white+black == 0 || white+black == 3 {
		return true
	}
	if white+black == 6 {
		var wB, bB uint64 = g.Bitboard.Pieces[4], g.Bitboard.Pieces[5]
		if wB != 0 && bB != 0 && wB&dark == bB&dark {
			return true
		}
	}
	return false
}
