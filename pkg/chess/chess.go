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
	// Move chan to handle incomming players' moves.
	Move chan bitboard.Move `json:"-"`
	// End channel is used by Room to be able to terminate the
	// Run goroutine in case both players are disconnected.
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
		Move:        make(chan bitboard.Move),
		End:         make(chan struct{}),
	}
	g.Bitboard.GenLegalMoves()
	return g
}

// Run handles incomming game events, such as players' moves and clock ticks.
// TODO: escape callbacks.
func (g *Game) Run(moveCallback func(m CompletedMove), timeoutCallback func()) {
	defer func() {
		g.Clock.Stop()
		log.Printf("clock stoped")
	}()

	for {
		select {
		case m := <-g.Move:
			if g.ProcessMove(m) {
				moveCallback(g.Moves[len(g.Moves)-1])
				if g.Result != enums.Unknown {
					return
				}
			}

		case <-g.Clock.C:
			if g.Bitboard.ActiveColor == enums.White {
				g.WhiteTime--
			} else {
				g.BlackTime--
			}

			if g.WhiteTime <= 0 {
				g.setEndInfo(enums.Timeout, enums.Black)
				timeoutCallback()
				return
			} else if g.BlackTime <= 0 {
				g.setEndInfo(enums.Timeout, enums.White)
				timeoutCallback()
				return
			}

		case <-g.End:
			return
		}
	}
}

func (g *Game) ProcessMove(m bitboard.Move) bool {
	if !g.Bitboard.IsMoveLegal(m) || g.Result != enums.Unknown {
		return false
	}

	c := g.Bitboard.ActiveColor
	movedPT := bitboard.GetPieceOnSquare(1<<m.From(), g.Bitboard.Pieces)
	piecesBefore := g.Bitboard.Pieces
	lm := g.Bitboard.LegalMoves[:]

	g.Bitboard.MakeMove(m)

	g.Bitboard.SetCastlingRights(movedPT)

	g.Bitboard.SetEPTarget(m)

	// Switch the active color
	g.Bitboard.ActiveColor ^= 1

	// Generate legal moves for the next color.
	g.Bitboard.GenLegalMoves()

	timeLeft := 0
	if c == enums.White {
		g.WhiteTime += g.TimeBonus
		timeLeft = g.WhiteTime
	} else {
		g.BlackTime += g.TimeBonus
		timeLeft = g.BlackTime
	}

	g.Bitboard.SetHalfmoveCnt(movedPT, m.Type())

	// After the black moves, increment the full move counter.
	if c == enums.Black {
		g.Bitboard.FullmoveCnt++
	}

	isCheck := bitboard.GenAttackedSquares(g.Bitboard.Pieces, c)&
		g.Bitboard.Pieces[10+g.Bitboard.ActiveColor] != 0

	isCheckmate := isCheck && len(g.Bitboard.LegalMoves) == 0
	g.Moves = append(g.Moves, CompletedMove{
		Move:     m,
		SAN:      san.Move2SAN(m, piecesBefore, lm, movedPT, isCheck, isCheckmate),
		FEN:      fen.Bitboard2FEN(g.Bitboard),
		TimeLeft: timeLeft,
	})

	if isCheckmate {
		g.setEndInfo(enums.Checkmate, c)
		return true
	} else if len(g.Bitboard.LegalMoves) == 0 {
		g.setEndInfo(enums.Stalemate, enums.None)
		return true
	}

	if g.isThreefoldRepetition() {
		g.setEndInfo(enums.Repetition, enums.None)
		return true
	}

	if g.isInsufficientMaterial() {
		g.setEndInfo(enums.InsufficientMaterial, enums.None)
		return true
	}

	if g.Bitboard.HalfmoveCnt == 100 {
		g.setEndInfo(enums.FiftyMoves, enums.None)
		return true
	}

	return true
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
	// Bitmask for all dark squares.
	var dark uint64 = 0xAA55AA55AA55AA55
	material := g.Bitboard.CalculateMaterial()

	if material == 0 {
		return true
	}

	if material == 3 && g.Bitboard.Pieces[enums.WhitePawn] == 0 &&
		g.Bitboard.Pieces[enums.BlackPawn] == 0 {
		return true
	}

	if material == 6 {
		wb, bb := g.Bitboard.Pieces[enums.WhiteBishop], g.Bitboard.Pieces[enums.BlackBishop]
		if wb != 0 && bb != 0 && ((wb&dark > 0 && bb&dark > 0) ||
			(wb&dark == 0 && bb&dark == 0)) {
			return true
		}
	}
	return false
}

func (g *Game) setEndInfo(r enums.Result, w enums.Color) {
	g.Result = r
	g.Winner = w
}
