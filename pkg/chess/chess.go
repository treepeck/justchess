// Package chess implements chess logic.
package chess

import (
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
	"justchess/pkg/chess/fen"
	"justchess/pkg/chess/san"
	"log"
	"time"

	"github.com/google/uuid"
)

type CompletedMove struct {
	Move bitboard.Move `json:"m"`
	SAN  string        `json:"s"`
	// Bitboard state after completing the move.
	FEN string `json:"f"`
	// Remaining time on a player's clock in seconds.
	TimeLeft int `json:"t"`
}

// Game represents a single chess game. All time values are stored in seconds.
type Game struct {
	// Used only in the database. Must be equal to the [ws.Room] id.
	Id         uuid.UUID          `json:"id"`
	Result     enums.Result       `json:"r"`
	Winner     enums.Color        `json:"w"`
	Bitboard   *bitboard.Bitboard `json:"-"`
	InitialFEN string             `json:"-"`
	Moves      []CompletedMove    `json:"m"`
	// Used only in the database.
	WhiteId uuid.UUID `json:"wid"`
	// Used only in the database.
	BlackId     uuid.UUID    `json:"bid"`
	WhiteTime   int          `json:"-"`
	BlackTime   int          `json:"-"`
	Clock       *time.Ticker `json:"-"`
	TimeControl int          `json:"tc"`
	TimeBonus   int          `json:"tb"`
	// End channel is used to terminate the DecrementTime goroutine when the
	// game was ended for the reason different from timeout.
	End chan struct{} `json:"-"`
}

func NewGame(id uuid.UUID, bb *bitboard.Bitboard, control, bonus int) *Game {
	if bb == nil {
		bb = fen.FEN2Bitboard("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	}
	g := &Game{
		Id:          id,
		Result:      enums.Unknown,
		Winner:      enums.None,
		Bitboard:    bb,
		InitialFEN:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Moves:       make([]CompletedMove, 0),
		WhiteTime:   control,
		BlackTime:   control,
		Clock:       time.NewTicker(time.Second), // The timer will send a signal each second.
		TimeBonus:   bonus,
		TimeControl: control,
		End:         make(chan struct{}),
	}
	g.Bitboard.GenLegalMoves()
	return g
}

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

		movedPT := bitboard.GetPieceOnSquare(1<<m.From(), g.Bitboard.Pieces)
		var SAN = san.Move2SAN(m, g.Bitboard.Pieces, g.Bitboard.LegalMoves, movedPT)
		g.Bitboard.MakeMove(legalMove)

		// Castling is no more possible if the king has moved, or the rooks are not on their standart
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
				g.Winner = c
				g.End <- struct{}{}
			} else {
				g.Result = enums.Stalemate
				g.Winner = enums.None
				g.End <- struct{}{}
			}
		}

		timeLeft := 0
		if c == enums.White {
			g.WhiteTime += g.TimeBonus
			timeLeft = g.WhiteTime
		} else {
			g.BlackTime += g.TimeBonus
			timeLeft = g.BlackTime
		}
		g.Moves = append(g.Moves, CompletedMove{
			Move:     m,
			SAN:      SAN,
			FEN:      fen.Bitboard2FEN(g.Bitboard),
			TimeLeft: timeLeft,
		})

		if g.isThreefoldRepetition() {
			g.Result = enums.Repetition
			g.Winner = enums.None
			g.End <- struct{}{}
		}

		if g.isInsufficientMaterial() {
			g.Result = enums.InsufficientMaterial
			g.Winner = enums.None
			g.End <- struct{}{}
		}

		if g.Bitboard.HalfmoveCnt == 100 {
			g.Result = enums.FiftyMoves
			g.Winner = enums.None
			g.End <- struct{}{}
		}

		return true
	}
	return false
}

// When one of the players runs out of time, the game calls the callback.
// The callback should be a method which notifies the players about the game result.
func (g *Game) DecrementTime(callback func()) {
	defer func() {
		g.Clock.Stop()
		log.Printf("clock stoped")
	}()

	for {
		select {
		// Wait for time ticks.
		case <-g.Clock.C:

			if g.Bitboard.ActiveColor == enums.White {
				g.WhiteTime--
			} else {
				g.BlackTime--
			}

			if g.WhiteTime <= 0 {
				g.Result = enums.Timeout
				g.Winner = enums.Black
				callback()
				return
			} else if g.BlackTime <= 0 {
				g.Result = enums.Timeout
				g.Winner = enums.White
				callback()
				return
			}

		case <-g.End:
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
	// Mask for all dark squares.
	var dark uint64 = 0xAA55AA55AA55AA55
	material := g.Bitboard.CalculateMaterial()

	// Case 1.
	if material == 0 {
		return true
	}
	// Case 2.
	if material == 3 && g.Bitboard.Pieces[enums.WhitePawn] == 0 &&
		g.Bitboard.Pieces[enums.BlackPawn] == 0 {
		return true
	}
	// Case 3.
	if material == 6 {
		wb, bb := g.Bitboard.Pieces[enums.WhiteBishop], g.Bitboard.Pieces[enums.BlackBishop]
		if wb != 0 && bb != 0 && ((wb&dark > 0 && bb&dark > 0) ||
			(wb&dark == 0 && bb&dark == 0)) {
			return true
		}
	}
	return false
}
