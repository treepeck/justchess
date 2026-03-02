package ws

import (
	"encoding/json"
	"justchess/internal/db"
	"log"
	"strings"
	"time"

	"github.com/treepeck/chego"
)

const (
	// Disconnected player has 30 seconds to reconnect.  If the player doesn't
	// reconnect within the specified time period, victory is awarded to the
	// other player if they are online.  If both players are disconnected the
	// game is marked as abandoned and will not be scored.
	reconnectDeadline int = 30

	// Minimal number of moves required to correctly complete the game
	// otherwise, the game will be marked as abandoned.
	minMoves = 3
)

type room struct {
	moves []completedMove
	id    string
	white db.Player
	black db.Player
	// Id of the player who have issues a draw offer.
	// Empty string in case no pending draw offers.
	drawOfferIssuer    string
	whiteReconnectTime int
	blackReconnectTime int
	// To calculate the amount of time it took to make a move.
	timeBeforeMove int
	clients        map[string]*client
	// When timeToLive is equal to 0, the room will destroy itself.
	register   chan *client
	unregister chan string
	handle     chan event
	store      chan<- storeGame
	clock      *time.Ticker
	game       *chego.Game
	// To avoid draw offer spamming limit the number of  draw offers
	// to 1 from each player.
	// TODO: reset after 10 moves.
	hasWhiteSentDrawOffer bool
	hasBlackSentDrawOffer bool
}

func newRoom(
	id string, white, black db.Player,
	control, bonus int, store chan<- storeGame,
) *room {
	g := chego.NewGame()
	g.SetClock(control*60, bonus)

	log.Printf("control: %d, bonus: %d", control, bonus)

	return &room{
		id:                 id,
		white:              white,
		black:              black,
		moves:              make([]completedMove, 0),
		game:               g,
		whiteReconnectTime: reconnectDeadline,
		blackReconnectTime: reconnectDeadline,
		timeBeforeMove:     control * 60,
		clients:            make(map[string]*client),
		register:           make(chan *client),
		unregister:         make(chan string),
		handle:             make(chan event),
		store:              store,
		clock:              time.NewTicker(time.Second),
	}
}

func (r *room) listenEvents(remove chan<- string) {
	defer func() { remove <- r.id }()

	for {
		select {
		case c := <-r.register:
			r.handleRegister(c)

		case id := <-r.unregister:
			r.handleUnregister(id)

		case e := <-r.handle:
			switch e.Action {
			case actionChat:
				r.handleChat(e)
			case actionMove:
				r.handleMove(e)
			case actionResign:
				r.handleResign(e)
			case actionOfferDraw:
				r.handleOfferDraw(e)
			case actionAcceptDraw:
				r.handleAcceptDraw(e)
			case actionDeclineDraw:
				r.handleDeclineDraw(e)
			}

		case <-r.clock.C:
			r.handleTimeTick()
			// Destroy the room if both players have been disconnected for a while.
			if r.whiteReconnectTime < 1 && r.blackReconnectTime < 1 {
				r.clock.Stop()
				return
			}
		}
	}
}

func (r *room) handleRegister(c *client) {
	// Decline the connection if the client is already in the queue.
	if _, exist := r.clients[c.player.Id]; exist {
		// Send error event to the client.
		if raw, err := newEncodedEvent(actionError, msgConflict); err == nil {
			c.send <- raw
		} else {
			log.Print(err)
		}
		return
	}

	log.Printf("client %s joined room %s", c.player.Id, r.id)

	c.unregister = r.unregister
	c.forward = r.handle
	r.clients[c.player.Id] = c
	// Send the game state so that the client can sync.
	raw, err := newEncodedEvent(actionGame, gamePayload{
		LegalMoves:  r.game.LegalMoves.Moves[:r.game.LegalMoves.LastMoveIndex],
		Moves:       r.moves,
		Termination: r.game.Termination,
		Result:      r.game.Result,
		WhiteTime:   r.game.WhiteTime,
		BlackTime:   r.game.BlackTime,
	})
	if err != nil {
		log.Print(err)
		return
	}
	c.send <- raw

	r.broadcast(actionConn, c.player.Name)
}

func (r *room) handleUnregister(id string) {
	c, exists := r.clients[id]
	if !exists {
		log.Printf("client is not registered")
		return
	}

	log.Printf("client %s leaves room %s", id, r.id)

	delete(r.clients, id)

	r.broadcast(actionDisc, c.player.Name)
}

// handleTimeTick decrements the time on active player's clock.
func (r *room) handleTimeTick() {
	// If some player is disconnected, decrement their allowed reconnect time.
	if _, isConnected := r.clients[r.white.Id]; !isConnected {
		r.whiteReconnectTime--
	}
	if _, isConnected := r.clients[r.black.Id]; !isConnected {
		r.blackReconnectTime--
	}

	// Shortcut: game is already over.
	if r.game.Result != chego.Unknown {
		return
	}

	// Terminate the game if one of the player failed to reconnect.
	if r.whiteReconnectTime < 1 || r.blackReconnectTime < 1 {
		if len(r.moves) < minMoves {
			// Mark game as abandoned if less then 2 moves were made.
			r.endGame(chego.Abandoned, chego.Unknown)
		} else if r.blackReconnectTime < 1 {
			r.endGame(chego.TimeForfeit, chego.WhiteWon)
		} else {
			r.endGame(chego.TimeForfeit, chego.BlackWon)
		}
		return
	}

	// Decrement player's clock time.
	if r.game.Position.ActiveColor == chego.ColorWhite {
		r.game.WhiteTime--
	} else {
		r.game.BlackTime--
	}

	// Terminate the game due to the time forfeit.
	if r.game.WhiteTime == 0 || r.game.BlackTime == 0 {
		if r.game.IsInsufficientMaterial() {
			// Accord to chess rules, the game is draw if opponent doesn't have
			// sufficient material to checkmate you.
			r.endGame(chego.TimeForfeit, chego.Draw)
		} else if len(r.moves) < minMoves {
			// Mark game as abandoned if less then 2 moves were made.
			r.endGame(chego.Abandoned, chego.Unknown)
		} else {
			if r.game.BlackTime == 0 {
				r.endGame(chego.TimeForfeit, chego.WhiteWon)
			} else {
				r.endGame(chego.TimeForfeit, chego.BlackWon)
			}
		}
	}
}

// handleMove validates, performes, stores, and broadcasts the move.
// Also ends the game if some endgame state is reached.
// The event will be ignored if the sender does not have the right to move
// or the game is already over.
func (r *room) handleMove(e event) {
	if (len(r.moves)%2 == 0 && e.sender.player.Id != r.white.Id) ||
		(len(r.moves)%2 != 0 && e.sender.player.Id != r.black.Id) ||
		r.game.Termination != chego.Unterminated {
		return
	}

	// Decline the move if it is not legal.
	var index byte
	err := json.Unmarshal(e.Payload, &index)
	if err != nil || index >= r.game.LegalMoves.LastMoveIndex {
		return
	}

	// Perform and store the move.
	m := r.game.LegalMoves.Moves[index]
	completed := completedMove{
		San:   r.game.PushMove(m),
		Fen:   chego.SerializeBitboards(r.game.Position.Bitboards),
		Move:  m,
		index: index,
	}
	// Store time after completing the move to synchronize clock on frontend.
	var timeLeft int
	if r.game.Position.ActiveColor == chego.ColorWhite {
		completed.timeDiff = r.timeBeforeMove - r.game.BlackTime
		r.timeBeforeMove = r.game.WhiteTime
		timeLeft = r.game.BlackTime
	} else {
		completed.timeDiff = r.timeBeforeMove - r.game.WhiteTime
		r.timeBeforeMove = r.game.BlackTime
		timeLeft = r.game.WhiteTime
	}
	r.moves = append(r.moves, completed)

	r.broadcast(actionMove, movePayload{
		LegalMoves: r.game.LegalMoves.Moves[:r.game.LegalMoves.LastMoveIndex],
		TimeLeft:   timeLeft,
		Move:       r.moves[len(r.moves)-1],
	})

	// End the game according to the rules of chess.
	if r.game.IsCheckmate() {
		if len(r.moves)%2 == 0 {
			r.endGame(chego.Checkmate, chego.BlackWon)
		} else {
			r.endGame(chego.Checkmate, chego.WhiteWon)
		}
	} else if r.game.IsInsufficientMaterial() {
		r.endGame(chego.InsufficientMaterial, chego.Draw)
	} else if r.game.IsThreefoldRepetition() {
		r.endGame(chego.ThreefoldRepetition, chego.Draw)
	} else if r.game.LegalMoves.LastMoveIndex == 0 {
		r.endGame(chego.Stalemate, chego.Draw)
	} else if r.game.Position.HalfmoveCnt == 50 {
		r.endGame(chego.FiftyMoves, chego.Draw)
	}
}

// handleChat append sender name and broadcasts the message.
// TODO: sanityze and rate limit messages.
func (r *room) handleChat(e event) {
	var b strings.Builder
	// Append sender's name.
	b.WriteString(e.sender.player.Name)
	b.WriteString(": ")
	// Append message.
	b.WriteString(strings.TrimSpace(strings.ReplaceAll(string(e.Payload), "\"", " ")))

	e.Payload = json.RawMessage(b.String())
	r.broadcast(actionChat, b.String())
}

// handleResign handles player resignation.  Resignation will be denied if one
// of the following is true:
//   - There were not enough moves played to end the game;
//   - The game is already over;
//   - Sender is not a white or black player.
func (r *room) handleResign(e event) {
	if len(r.moves) < minMoves || r.game.Termination != chego.Unterminated {
		return
	}
	if r.white.Id == e.sender.player.Id {
		r.endGame(chego.Resignation, chego.BlackWon)
	} else if r.black.Id == e.sender.player.Id {
		r.endGame(chego.Resignation, chego.WhiteWon)
	}
}

// handleOfferDraw handles draw offers. Event will be denied if one
// of the following is true:
//   - There were not enough moves played to end the game;
//   - The game is already over;
//   - Sender is not a white nor a black player;
//   - One of the players has already sent a pending draw offer;
//   - The player has sent a draw offer not so long ago.
func (r *room) handleOfferDraw(e event) {
	if len(r.moves) < minMoves || r.game.Termination != chego.Unterminated ||
		(e.sender.player.Id != r.white.Id && e.sender.player.Id != r.black.Id) ||
		(e.sender.player.Id == r.white.Id && r.hasWhiteSentDrawOffer) ||
		(e.sender.player.Id == r.black.Id && r.hasBlackSentDrawOffer) ||
		r.drawOfferIssuer != "" {
		return
	}

	r.drawOfferIssuer = e.sender.player.Id
	// Send draw offer confirmation event to opponent.
	switch r.drawOfferIssuer {
	case r.white.Id:
		r.hasWhiteSentDrawOffer = true
		if c, isConnected := r.clients[r.black.Id]; isConnected {
			raw, err := newEncodedEvent(actionOfferDraw, nil)
			if err == nil {
				c.send <- raw
			}
		}
	case r.black.Id:
		r.hasBlackSentDrawOffer = true
		if c, isConnected := r.clients[r.white.Id]; isConnected {
			raw, err := newEncodedEvent(actionOfferDraw, nil)
			if err == nil {
				c.send <- raw
			}
		}
	}

	// Broadcast draw offer chat message.
	r.broadcast(actionChat, e.sender.player.Name+" offered draw")
}

// handleAcceptDraw accepts the draw offer.
func (r *room) handleAcceptDraw(e event) {
	if r.drawOfferIssuer == "" || e.sender.player.Id == r.drawOfferIssuer ||
		len(r.moves) < minMoves {
		return
	}
	r.drawOfferIssuer = ""
	r.endGame(chego.Agreement, chego.Draw)
}

// handleDeclineDraw declines the draw offer.
func (r *room) handleDeclineDraw(e event) {
	if r.drawOfferIssuer == "" || e.sender.player.Id == r.drawOfferIssuer ||
		len(r.moves) < minMoves {
		return
	}
	r.drawOfferIssuer = ""
	r.broadcast(actionChat, e.sender.player.Name+" declined draw")
}

// Sets the game termination and results and broadcasts [actionEnd] event.
// Writes the game into a database.
func (r *room) endGame(t chego.Termination, res chego.Result) {
	r.game.Termination = t
	r.game.Result = res
	r.broadcast(actionEnd, endPayload{Termination: t, Result: res})
	r.store <- storeGame{
		white: r.white, black: r.black,
		moves: r.moves, id: r.id,
		result: r.game.Result, termination: r.game.Termination,
	}
}

// broadcast encodes and sends the event to all connected clients.
func (r *room) broadcast(a eventAction, payload any) {
	raw, err := newEncodedEvent(a, payload)
	if err != nil {
		log.Print(err)
		return
	}

	for _, c := range r.clients {
		c.send <- raw
	}
}
