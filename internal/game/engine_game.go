package game

import (
	"justchess/internal/db"
	"log"

	"github.com/treepeck/chego"
)

type EngineGame struct {
	chego.Game

	// Indices of played moves for Huffman decoding.
	playedIndices   []byte
	id              string
	playerId        string
	gameRepo        db.GameRepo
	playerColor     chego.Color
	playerReconnect int
	isPlayerOnline  bool
}

// SpawnEngineGame inserts a new engine game record into repository and initializes
// [EngineGame] fields.
func SpawnEngineGame(id, playerId string, c chego.Color, gr db.GameRepo) (*EngineGame, error) {
	err := gr.InsertEngine(id, playerId, c)
	if err != nil {
		return nil, err
	}
	return &EngineGame{
		Game:            chego.NewGame(),
		id:              id,
		playerId:        playerId,
		playedIndices:   make([]byte, 0),
		gameRepo:        gr,
		playerColor:     c,
		playerReconnect: reconnectDeadline,
	}, nil
}

func (g *EngineGame) Play(id string, index byte) (MovePayload, bool) {
	if id != g.playerId || g.Termination != chego.Unterminated ||
		index >= g.Legal.LastMoveIndex {
		return MovePayload{}, false
	}

	m := g.Legal.Moves[index]
	g.Push(m)
	g.playedIndices = append(g.playedIndices, index)

	if g.Termination != chego.Unterminated {
		g.store()
	}

	return MovePayload{
		Legal:      g.Legal.Moves[:g.Legal.LastMoveIndex],
		PlayedMove: g.Played[len(g.Played)-1],
		Move:       m,
	}, true
}

func (g *EngineGame) Join(id string) {
	if id == g.playerId {
		g.isPlayerOnline = true
	}
	log.Printf("player %s joins game %s", id, g.id)
}

func (g *EngineGame) Leave(id string) {
	if id == g.playerId {
		g.isPlayerOnline = false
	}
	log.Printf("player %s leaves game %s", id, g.id)
}

func (g *EngineGame) TimeTick() {
	if g.Termination != chego.Unterminated {
		return
	}

	if !g.isPlayerOnline {
		g.playerReconnect--
	}

	if g.playerReconnect == 0 {
		g.Abandon()
	}
}

func (g *EngineGame) Resign(id string) bool {
	if len(g.Played) < minMoves || g.Termination != chego.Unterminated ||
		id != g.playerId {
		return false
	}
	r := chego.WhiteWon
	if g.playerColor == chego.ColorWhite {
		r = chego.BlackWon
	}
	g.Terminate(chego.Resignation, r)
	g.store()
	return true
}

func (g *EngineGame) Abandon() {
	if g.Termination == chego.Unterminated {
		g.Termination = chego.Abandoned
		if err := g.gameRepo.MarkEngineAsAbandoned(g.id); err != nil {
			log.Print(err)
		}
	}
}

func (g *EngineGame) store() {
	if err := g.gameRepo.UpdateEngine(db.EngineGameUpdate{
		Id: g.id, Result: g.Result, Termination: g.Termination,
		EncodedMoves: chego.HuffmanEncoding(g.playedIndices),
		MovesLength:  len(g.Played),
	}); err != nil {
		log.Print(err)
		return
	}
}

func (g *EngineGame) EndPayload() EndPayload {
	return EndPayload{
		Result:      g.Result,
		Termination: g.Termination,
	}
}

func (g *EngineGame) GamePayload() GamePayload {
	return GamePayload{
		Legal:  g.Legal.Moves[:g.Legal.LastMoveIndex],
		Played: g.Played,
	}
}
