package game

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"justchess/pkg/game/fen"
	"justchess/pkg/game/san"
	"time"

	"github.com/google/uuid"
)

type CompletedMove struct {
	SAN string
	// Biboard state after completing the move.
	FEN string
}

type Game struct {
	Result      enums.Result
	Bitboard    *bitboard.Bitboard
	Moves       []CompletedMove
	WhiteId     uuid.UUID
	BlackId     uuid.UUID
	WhiteTime   uint8
	BlackTime   uint8
	Timer       *time.Ticker
	TimeControl uint8
	TimeBonus   uint8
}

func NewGame(r enums.Result, bb *bitboard.Bitboard, control, bonus uint8) *Game {
	if bb == nil {
		bb = bitboard.NewBitboard([12]uint64{0xFF00, 0xFF000000000000, 0x42,
			0x4200000000000000, 0x24, 0x2400000000000000, 0x81, 0x8100000000000000,
			0x8, 0x800000000000000, 0x10, 0x1000000000000000}, enums.White,
			[4]bool{true, true, true, true}, -1, 0, 0)
	}
	g := &Game{
		Result:      r,
		Bitboard:    bb,
		Moves:       make([]CompletedMove, 0),
		WhiteTime:   control,
		BlackTime:   control,
		Timer:       time.NewTicker(time.Second), // The timer will send a signal each second.
		TimeBonus:   bonus,
		TimeControl: control,
	}
	return g
}

// ProcessMove processes only legal moves and returns true if the move was processed,
// false otherwise.
func (g *Game) ProcessMove(m bitboard.Move) bool {
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
		} else if m.Type() == enums.DoublePawnPush {
			if g.Bitboard.ActiveColor == enums.White {
				g.Bitboard.EPTarget = m.From() + 8
			} else {
				g.Bitboard.EPTarget = m.From() - 8
			}
		}
		ptBefore := bitboard.GetPieceTypeFromSquare(m.From(), g.Bitboard.Pieces)
		san := san.Move2SAN(legalMove, g.Bitboard.Pieces, g.Bitboard.LegalMoves, ptBefore)
		g.Bitboard.MakeMove(legalMove, ptBefore)
		c, opC := g.Bitboard.ActiveColor, g.Bitboard.ActiveColor^1
		g.Bitboard.ActiveColor = opC

		if g.isThreefoldRepetition() {
			g.Result = enums.Repetition
			return true
		}

		if g.isInsufficientMaterial() {
			g.Result = enums.InsufficienMaterial
			return true
		}
		// In case of promotion, the piece type will change.
		newPT := bitboard.GetPieceTypeFromSquare(m.To(), g.Bitboard.Pieces)
		// TRICK: Store the current castling rights.
		// The checked king will not be able to castle on the next move,
		// but should be able to castle later if the king and rooks did not make moves.
		crCopy := make([]bool, 4)
		copy(crCopy[:], g.Bitboard.CastlingRights[:])
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
		copy(g.Bitboard.CastlingRights[:], crCopy[:])
		if len(g.Bitboard.LegalMoves) == 0 {
			if isCheck {
				g.Result = enums.Checkmate
				san = san[:len(san)-1] + "#"
			} else {
				g.Result = enums.Stalemate
			}
		}
		g.Moves = append(g.Moves, CompletedMove{
			SAN: san,
			FEN: fen.Bitboard2FEN(g.Bitboard),
		})
		return true
	}
	return false
}

func (g *Game) DecrementTime() {
	for {
		// Wait for time ticks.
		_, ok := <-g.Timer.C
		if !ok {
			return
		}

		if g.Bitboard.ActiveColor == enums.White {
			g.WhiteTime--
		} else {
			g.BlackTime--
		}
		if g.WhiteTime <= 0 || g.BlackTime <= 0 {
			g.Timer.Stop()
			g.Result = enums.Timeout
			return
		}
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
	material := g.Bitboard.CalculateMaterial()

	if material == 0 || material == 3 {
		return true
	} else if material == 6 {
		var wB, bB uint64 = g.Bitboard.Pieces[4], g.Bitboard.Pieces[5]
		if wB != 0 && bB != 0 && wB&dark == bB&dark {
			return true
		}
	}
	return false
}
