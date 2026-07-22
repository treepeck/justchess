package game

import (
	"justchess/internal/db"
	"justchess/internal/talk"
)

// Storage serves as in-memory storage of the [Game] states.
type Storage struct {
	Create   chan talk.GameCreator
	Find     chan talk.GameFinder
	gameRepo db.GameRepo
	games    map[string]*game
}

func NewStorage(gr db.GameRepo) Storage {
	return Storage{
		Create:   make(chan talk.GameCreator),
		Find:     make(chan talk.GameFinder),
		gameRepo: gr,
		games:    make(map[string]*game),
	}
}

func (s Storage) Listen() {
	for {
		select {
		case c := <-s.Create:
			s.onCreate(c)
		case f := <-s.Find:
			s.onFind(f)
		}
	}
}

func (s Storage) onCreate(c talk.GameCreator) {
	g := db.Game{
		Id: c.Id,
		White: db.Player{
			Id: c.WhiteId,
		},
		Black: db.Player{
			Id: c.BlackId,
		},
		TimeControl: c.TimeControl,
		TimeBonus:   c.TimeBonus,
	}
	err := s.gameRepo.Insert(g)
	if err != nil {
		s.games[c.Id] = newGame(g)
	}
	c.Res <- err
}

func (s Storage) onFind(f talk.GameFinder) {
	// TODO: add db access layer. If game was not found in RAM,
	// fetch the DB and try to restore the game from last-known state.
	g, exists := s.games[f.Id]
	if !exists {
		f.Res <- talk.GameChannels{In: nil, Out: nil, Ban: nil}
	}
	f.Res <- g.channels
}
