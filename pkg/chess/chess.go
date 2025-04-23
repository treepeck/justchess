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

type Game struct {
	Result      enums.Result
	Winner      enums.Color
	Bitboard    *bitboard.Bitboard
	WhiteTime   int
	BlackTime   int
	WhiteId     uuid.UUID
	BlackId     uuid.UUID
	TimeControl int
	TimeBonus   int
	Moves       []CompletedMove
	clock       *time.Ticker
	// To handle incomming players' moves.
	Move chan MoveEvent
	// To safely get game info.
	Info chan GameInfoEvent
	// End channel is used by Room to be able to terminate the
	// RunRoutine if either both players are disconnected, one of them resigns,
	// or they both agreeded to a draw.
	End chan EndGameInfo
}

func NewGame(fenStr string, control, bonus int) *Game {
	var bb *bitboard.Bitboard
	if fenStr == "" {
		bb = fen.DefaultBB
	} else {
		bb = fen.FEN2Bitboard(fenStr)
	}

	g := &Game{
		Result:      enums.Unknown,
		Winner:      enums.None,
		Bitboard:    bb,
		Moves:       make([]CompletedMove, 0),
		WhiteTime:   control,
		BlackTime:   control,
		clock:       time.NewTicker(time.Second),
		TimeControl: control,
		TimeBonus:   bonus,
		Move:        make(chan MoveEvent),
		Info:        make(chan GameInfoEvent),
		End:         make(chan EndGameInfo),
	}

	g.Bitboard.GenLegalMoves()

	return g
}

func (g *Game) RunRoutine(timeout chan<- struct{}) {
	defer func() {
		g.clock.Stop()
		log.Printf("clock stoped")
	}()

	for {
		select {
		case me := <-g.Move:
			me.Response <- g.ProcessMove(me.Move)
			if g.Result != enums.Unknown {
				return
			}

		case <-g.clock.C:
			g.handleTimeTick()

			if g.WhiteTime == 0 {
				g.SetEndInfo(enums.Timeout, enums.Black)
				timeout <- struct{}{}
				return
			} else if g.BlackTime == 0 {
				g.SetEndInfo(enums.Timeout, enums.White)
				timeout <- struct{}{}
				return
			}

		case gie := <-g.Info:
			gie.Response <- GameInfo{
				WhiteTime: g.WhiteTime, BlackTime: g.BlackTime,
				Result: g.Result, Winner: g.Winner, Moves: g.Moves[:],
				LegalMoves: g.Bitboard.LegalMoves[:],
			}

		case info := <-g.End:
			g.Result = info.Result
			g.Winner = info.Winner
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
		g.SetEndInfo(enums.Checkmate, c)
		return true
	} else if len(g.Bitboard.LegalMoves) == 0 {
		g.SetEndInfo(enums.Stalemate, enums.None)
		return true
	}

	if g.isThreefoldRepetition() {
		g.SetEndInfo(enums.Repetition, enums.None)
		return true
	}

	if g.isInsufficientMaterial() {
		g.SetEndInfo(enums.InsufficientMaterial, enums.None)
		return true
	}

	if g.Bitboard.HalfmoveCnt == 100 {
		g.SetEndInfo(enums.FiftyMoves, enums.None)
		return true
	}

	return true
}

func (g *Game) handleTimeTick() {
	if g.Bitboard.ActiveColor == enums.White {
		g.WhiteTime--
	} else {
		g.BlackTime--
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

func (g *Game) SetEndInfo(r enums.Result, w enums.Color) {
	g.Result = r
	g.Winner = w
	g.clock.Stop()
}

type CompletedMove struct {
	Move     bitboard.Move `json:"m"`
	SAN      string        `json:"s"`
	FEN      string        `json:"f"` // Bitboard state after completing the move.
	TimeLeft int           `json:"t"` // Remaining time on a player's clock in seconds.
}

type MoveEvent struct {
	ClientId uuid.UUID
	Move     bitboard.Move
	Response chan<- bool // Whether the move was processed.
}

type GameInfo struct {
	WhiteTime  int
	BlackTime  int
	Result     enums.Result
	Winner     enums.Color
	Moves      []CompletedMove
	LegalMoves []bitboard.Move
}

type GameInfoEvent struct {
	Response chan<- GameInfo
}

type EndGameInfo struct {
	Result enums.Result
	Winner enums.Color
}
