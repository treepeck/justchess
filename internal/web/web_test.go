package web

import (
	"errors"
	"justchess/internal/db"
	"testing"

	"github.com/treepeck/chego"
)

type mockPlayerRepo struct {
}

func (r mockPlayerRepo) SelectById(id string) (db.Player, error) {
	return db.Player{}, errors.New("not implemented")
}

func (r mockPlayerRepo) SelectProfileData(id string) (db.ProfileData, error) {
	return db.ProfileData{}, errors.New("not implemented")
}

func (r mockPlayerRepo) SelectLeaderboard() ([]db.ProfileData, error) {
	return nil, errors.New("not implemented")
}

func (r mockPlayerRepo) SelectBySessionId(id string) (db.Player, error) {
	return db.Player{}, errors.New("not implemented")
}

func (r mockPlayerRepo) UpdateRatings(white, black db.RatingUpdate) error {
	return errors.New("not implemented")
}

type mockGameRepo struct {
}

func (r mockGameRepo) InsertRated(id, whiteId, blackId string, control, bonus int) error {
	return nil
}

func (r mockGameRepo) SelectRated(id string) (db.RatedGame, error) {
	return db.RatedGame{}, nil
}

func (r mockGameRepo) SelectNewestRated(id string) ([]db.RatedGameBrief, error) {
	return nil, nil
}

func (r mockGameRepo) SelectOlderRated(id string, p db.Pagination) ([]db.RatedGameBrief, error) {
	return nil, nil
}

func (r mockGameRepo) UpdateRated(gu db.RatedGameUpdate) error {
	return nil
}

func (r mockGameRepo) MarkRatedAsAbandoned(id string) error {
	return nil
}

func (r mockGameRepo) InsertEngine(id, playerId string, playerColor chego.Color) error {
	return nil
}

func (r mockGameRepo) SelectEngine(id string) (db.EngineGame, error) {
	return db.EngineGame{}, nil
}

func (r mockGameRepo) SelectNewestEngine(id string) ([]db.EngineGameBrief, error) {
	return nil, nil
}

func (r mockGameRepo) SelectOlderEngine(id string, p db.Pagination) ([]db.EngineGameBrief, error) {
	return nil, nil
}

func (r mockGameRepo) UpdateEngine(gu db.EngineGameUpdate) error {
	return nil
}

func (r mockGameRepo) MarkEngineAsAbandoned(id string) error {
	return nil
}

func BenchmarkParsePages(b *testing.B) {
	s := NewService(mockPlayerRepo{}, mockGameRepo{})

	for b.Loop() {
		s.ParsePages("../../_web/templates/")
	}
}
