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
	g.Bitboard.GenLegalMoves()
	return g
}

// TODO: refactoring, ProcessMove is too large.
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
		}
		c := g.Bitboard.ActiveColor
		if c == enums.Black {
			// After the black moves, the fullmove counter increments.
			g.Bitboard.FullmoveCnt++
		}

		movedPT := bitboard.GetPieceTypeFromSquare(1<<m.From(), g.Bitboard.Pieces)
		g.Bitboard.MakeMove(legalMove)

		// Castling is no more possible if the king has moved, or the rooks are not in their standart
		// positions.
		if movedPT == enums.WhiteKing || movedPT == enums.BlackKing {
			g.Bitboard.CastlingRights[0+c] = false
			g.Bitboard.CastlingRights[2+c] = false
		}
		if g.Bitboard.Pieces[enums.WhiteRook]&0x1 == 0 {
			g.Bitboard.CastlingRights[2] = false
		}
		if g.Bitboard.Pieces[enums.WhiteRook]&0x80 == 0 {
			g.Bitboard.CastlingRights[0] = false
		}
		if g.Bitboard.Pieces[enums.BlackRook]&0x100000000000000 == 0 {
			g.Bitboard.CastlingRights[3] = false
		}
		if g.Bitboard.Pieces[enums.BlackRook]&0x8000000000000000 == 0 {
			g.Bitboard.CastlingRights[1] = false
		}

		// Reset the en passant target since the en passant capture is possible only
		// for 1 move.
		g.Bitboard.EPTarget = enums.NoSquare

		switch m.Type() {
		// After double pawn push, set the en passant target.
		case enums.DoublePawnPush:
			if c == enums.White {
				g.Bitboard.EPTarget = m.To() - 8
			} else {
				g.Bitboard.EPTarget = m.To() + 8
			}

		// After altering material, the halfmove counter resets.
		case enums.Capture, enums.KnightPromo, enums.BishopPromo, enums.RookPromo,
			enums.QueenPromo, enums.KnightPromoCapture, enums.BishopPromoCapture,
			enums.RookPromoCapture, enums.QueenPromoCapture:
			g.Bitboard.HalfmoveCnt = 0

		// Increment halfmove counter if the move is not a capture or a pawn advance.
		default:
			if movedPT != enums.WhitePawn && movedPT != enums.BlackPawn {
				g.Bitboard.HalfmoveCnt++
			}
		}

		// Switch the active color
		g.Bitboard.ActiveColor ^= 1

		var SAN = san.Move2SAN(m, g.Bitboard.Pieces, g.Bitboard.LegalMoves, movedPT)
		// To determine if the last m was a check, generate possible moves
		// for the moved piece.
		var occupied uint64
		for _, board := range g.Bitboard.Pieces {
			occupied |= board
		}

		isCheck := bitboard.GenAttackedSquares(g.Bitboard.Pieces, c)&
			g.Bitboard.Pieces[10+g.Bitboard.ActiveColor] != 0
		if isCheck {
			SAN += "+"
		}

		// Generate legal moves for the next color.
		g.Bitboard.GenLegalMoves()
		if len(g.Bitboard.LegalMoves) == 0 {
			if isCheck {
				g.Result = enums.Checkmate
				SAN = SAN[:len(SAN)-1] + "#"
			} else {
				g.Result = enums.Stalemate
			}
		}
		g.Moves = append(g.Moves, CompletedMove{
			SAN: SAN,
			FEN: fen.Bitboard2FEN(g.Bitboard),
		})
		if g.isThreefoldRepetition() {
			g.Result = enums.Repetition
		}

		if g.isInsufficientMaterial() {
			g.Result = enums.InsufficienMaterial
		}
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
		var wB, bB uint64 = g.Bitboard.Pieces[enums.WhiteBishop],
			g.Bitboard.Pieces[enums.BlackBishop]
		if wB != 0 && bB != 0 && wB&dark == bB&dark {
			return true
		}
	}
	return false
}
