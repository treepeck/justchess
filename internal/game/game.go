package game

import (
	"encoding/json"
	"github.com/treepeck/chego"
	"justchess/internal/compression"
	"justchess/internal/db"
	"justchess/internal/talk"
	"log"
	"time"
)

const (
	// Draw can be offered only once in the specified amount of moves.
	drawOfferInterval = 10

	// Minimal number of moves that must be played to terminate the game.
	// Otherwise the game will be marked as abandoned.
	movesThreshold = 4

	errActionDeclined = "You have no rights to perform this action"
)

// game manages the state of a single arbitrary game. It is needed since the chego
// module implements pure chess logic, without any player and time related stuff.
type game struct {
	State                    db.Game
	moves                    []compression.Move
	playedIndices            []int
	channels                 talk.GameChannels
	id                       string
	position                 *chego.Position
	legal                    *chego.MoveList
	ticker                   *time.Ticker
	repetitions              map[uint64]int
	whiteTime                int
	blackTime                int
	timeBeforeMove           int
	movesSinceWhiteDrawOffer int
	movesSinceBlackDrawOffer int
	isPendingWhiteDrawOffer  bool
	isPendingBlackDrawOffer  bool
}

func newGame(id string, s db.Game) *game {
	return &game{
		State: s,
		channels: talk.GameChannels{
			In:  make(chan talk.Message, 192),
			Out: make(chan []byte),
			Ban: make(chan string),
		},
		ticker:         time.NewTicker(time.Second),
		moves:          make([]compression.Move, 0),
		playedIndices:  make([]int, 0),
		position:       chego.ParseFen(chego.InitialPos),
		legal:          &chego.MoveList{},
		repetitions:    make(map[uint64]int),
		whiteTime:      s.TimeControl,
		blackTime:      s.TimeControl,
		timeBeforeMove: s.TimeControl,
	}
}

func (g *game) MessagePump() {
	for {
		select {
		case m := <-g.channels.In:
			switch m.Kind {
			case talk.MessageMove:
				var moveIndex int
				if err := json.Unmarshal(m.Payload, &moveIndex); err != nil {
					g.channels.Ban <- m.PlayerId
					continue
				}
				g.play(m.PlayerId, moveIndex)
			case talk.MessageOfferDraw:
				g.offerDraw(m.PlayerId)
			case talk.MessageDeclineDraw:
				g.declineDraw(m.PlayerId)
			case talk.MessageAcceptDraw:
				g.acceptDraw(m.PlayerId)
			default:
				g.channels.Ban <- m.PlayerId
			}
		case <-g.ticker.C:
			g.timeTick()
		}
	}
}

func (g *game) play(playerId string, moveIndex int) {
	// Validate the move.
	isGameContinues := g.State.Termination == db.Unterminated
	isPlayerTurn := ((g.position.ActiveColor == chego.ColorWhite) && (playerId == g.State.White.Id)) ||
		((g.position.ActiveColor == chego.ColorBlack) && (playerId == g.State.Black.Id))
	isValidMove := moveIndex >= 0 || moveIndex < int(g.legal.Len)
	if !isGameContinues || !isPlayerTurn || !isValidMove {
		return
	}

	m := g.legal.Moves[moveIndex]
	moved := g.position.GetPieceFromSquare(1 << m.From())
	captured := g.position.GetPieceFromSquare(1 << m.To())
	isCapture := captured != chego.PieceNone

	// Move2SAN updates the position and generates legal moves for next turn.
	san := chego.Move2SAN(m, g.position, g.legal)

	// Clear the repetitions map after applying the irreversable move.
	// See https://www.chessprogramming.org/Irreversible_Moves
	if isCapture || m.Type() == chego.MoveCastling || m.Type() == chego.MovePromotion ||
		moved <= chego.BPawn {
		clear(g.repetitions)
	}

	// Increment the repitition key entry.
	g.repetitions[g.position.ZobristKey()]++

	// Store played move.
	g.State.Moves = append(g.State.Moves, compression.Move{
		San: san,
		Fen: chego.SerializeFen(g.position),
	})

	// Store time after completing the move to synchronize clock on frontend.
	var timeDiff, timeLeft int
	if g.State.TimeControl != 0 {
		if g.position.ActiveColor == chego.ColorWhite {
			g.blackTime += g.State.TimeBonus
			timeDiff = g.blackTime - g.timeBeforeMove
			g.timeBeforeMove = g.whiteTime
			timeLeft = g.blackTime
		} else {
			g.whiteTime += g.State.TimeBonus
			timeDiff = g.whiteTime - g.timeBeforeMove
			g.timeBeforeMove = g.blackTime
			timeLeft = g.whiteTime
		}
		g.State.TimeDiffs = append(g.State.TimeDiffs, timeDiff)
	}

	g.playedIndices = append(g.playedIndices, moveIndex)

	// Terminate the game according to the rules of chess.
	if g.isCheckmate() {
		if len(g.State.Moves)%2 == 0 {
			g.terminate(db.Checkmate, db.BlackWon)
		} else {
			g.terminate(db.Checkmate, db.WhiteWon)
		}
	} else if g.position.IsInsufficientMaterial() {
		g.terminate(db.InsufficientMaterial, db.Draw)
	} else if g.isThreefoldRepetition() {
		g.terminate(db.ThreefoldRepetition, db.Draw)
	} else if g.legal.Len == 0 {
		g.terminate(db.Stalemate, db.Draw)
	} else if g.position.HalfmoveCnt == 50 {
		g.terminate(db.FiftyMoves, db.Draw)
	}
	msg, err := talk.JSON(talk.MessageMove, Move{
		Fen:         chego.SerializeFen(g.position),
		San:         san,
		Result:      g.State.Result,
		Termination: g.State.Termination,
		Move:        int(m),
		TimeLeft:    timeLeft,
	})
	if err != nil {
		log.Print(err)
		return
	}
	g.channels.Out <- msg
}

func (g *game) offerDraw(playerId string) {
	if g.State.Termination != db.Unterminated || len(g.State.Moves) < movesThreshold {
		return
	}
	switch playerId {
	case g.State.White.Id:
		if g.movesSinceWhiteDrawOffer < drawOfferInterval || g.isPendingWhiteDrawOffer {
			return
		}
		g.movesSinceWhiteDrawOffer = 0
		g.isPendingWhiteDrawOffer = true
	case g.State.Black.Id:
		if g.movesSinceBlackDrawOffer < drawOfferInterval || g.isPendingBlackDrawOffer {
			return
		}
		g.movesSinceBlackDrawOffer = 0
		g.isPendingBlackDrawOffer = true
	default:
		g.channels.Ban <- playerId
		return
	}
	msg, err := talk.JSON(talk.MessageOfferDraw, playerId)
	if err != nil {
		log.Print(err)
		return
	}
	g.channels.Out <- msg
}

func (g *game) declineDraw(playerId string) {
	if g.State.Termination != db.Unterminated {
		return
	}
	if playerId == g.State.White.Id && g.isPendingBlackDrawOffer {
		g.isPendingBlackDrawOffer = false
	} else if playerId == g.State.Black.Id && g.isPendingWhiteDrawOffer {
		g.isPendingWhiteDrawOffer = false
	} else {
		g.channels.Ban <- playerId
		return
	}
	msg, err := talk.JSON(talk.MessageDeclineDraw, playerId)
	if err != nil {
		log.Print(err)
		return
	}
	g.channels.Out <- msg
}

func (g *game) acceptDraw(playerId string) {
	if g.State.Termination != db.Unterminated {
		return
	}
	if (playerId == g.State.White.Id && g.isPendingBlackDrawOffer) ||
		(playerId == g.State.Black.Id && g.isPendingWhiteDrawOffer) {
		g.terminate(db.Agreement, db.Draw)
		msg, err := talk.JSON(talk.MessageAcceptDraw, playerId)
		if err != nil {
			log.Print(err)
			return
		}
		g.channels.Out <- msg
	} else {
		g.channels.Ban <- playerId
	}
}

func (g *game) resign(playerId string) {
	if g.State.Termination != db.Unterminated || len(g.State.Moves) < movesThreshold {
		return
	}
	switch playerId {
	case g.State.White.Id:
		g.terminate(db.Resignation, db.BlackWon)
	case g.State.Black.Id:
		g.terminate(db.Resignation, db.WhiteWon)
	default:
		g.channels.Ban <- playerId
		return
	}
	msg, err := talk.JSON(talk.MessageResign, playerId)
	if err != nil {
		log.Print(err)
		return
	}
	g.channels.Out <- msg
}

// timeTick triggers every second to decrement the remaining time of player with active color.
func (g *game) timeTick() {
	if g.State.TimeControl == 0 || g.State.Termination != db.Unterminated {
		return
	}
	if g.position.ActiveColor == chego.ColorWhite {
		g.whiteTime--
	} else {
		g.blackTime--
	}
}

func (g *game) terminate(t db.Termination, r db.Result) {
	g.State.Termination = t
	g.State.Result = r
}

func (g *game) isCheckmate() bool {
	return chego.GenChecksCounter(g.position.Bitboards, 1^g.position.ActiveColor) > 0 &&
		g.legal.Len == 0
}

// isThreefoldRepetition checks whether the game has reached a threefold repetition.
//
// Two positions are considered identical if all of the following conditions are met:
//   - Active colors are the same.
//   - Pieces occupy the same squares.
//   - Legal moves are the same.
//   - Castling rights are identical.
//
// NOTE: Positions are identical even if the en passant target square differs,
// provided that no en passant capture is possible.
func (g *game) isThreefoldRepetition() bool {
	for _, numOfReps := range g.repetitions {
		if numOfReps >= 3 {
			return true
		}
	}
	return false
}
