// Package api implements HTTP API that serves database data.
package api

import (
	"encoding/json"
	"justchess/internal/db"
	"log"
	"net/http"
	"time"
)

// Declaration of error messages.
const (
	msgNotFound     = "The requested game wasn't found"
	msgBadRequest   = "Missing or invalid parameters"
	msgEncoderError = "Couldn't encode the response"
)

type Service struct {
	gameRepo db.SQLGameRepo
}

func NewService(gr db.SQLGameRepo) Service {
	return Service{
		gameRepo: gr,
	}
}

func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/rated-brief", s.ratedBrief)
	mux.HandleFunc("GET /api/engine-brief", s.engineBrief)
}

// ratedBrief serves brief info about the player's rated games.
func (s Service) ratedBrief(rw http.ResponseWriter, r *http.Request) {
	playerId := r.URL.Query().Get("pid")
	if len(playerId) != 12 {
		http.Error(rw, msgBadRequest, http.StatusBadRequest)
		return
	}
	// Optional cursor-based pagination parameters.
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cca"))

	var games []db.RatedGameBrief
	// If optional pagination paramers are defined.
	if err == nil && len(cursorId) == 12 {
		games, err = s.gameRepo.SelectRatedBrief(playerId, db.Pagination{
			CursorCreatedAt: cursorCreatedAt,
			CursorId:        cursorId,
		})
	} else {
		games, err = s.gameRepo.SelectLatestRatedBrief(playerId)
	}

	if err != nil {
		http.Error(rw, msgNotFound, http.StatusNotFound)
		return
	}

	if err = json.NewEncoder(rw).Encode(games); err != nil {
		log.Print(err)
		http.Error(rw, msgNotFound, http.StatusInternalServerError)
	}
}

// engineBrief serves brief info about player's engine games.
func (s Service) engineBrief(rw http.ResponseWriter, r *http.Request) {
	playerId := r.URL.Query().Get("pid")
	if len(playerId) != 12 {
		http.Error(rw, msgBadRequest, http.StatusBadRequest)
		return
	}
	// Optional cursor-based pagination parameters.
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cca"))

	var games []db.EngineGameBrief
	// If optional pagination paramers are defined.
	if err == nil && len(cursorId) == 12 {
		games, err = s.gameRepo.SelectEngineBrief(playerId, db.Pagination{
			CursorCreatedAt: cursorCreatedAt,
			CursorId:        cursorId,
		})
	} else {
		games, err = s.gameRepo.SelectLatestEngineBrief(playerId)
	}

	if err != nil {
		http.Error(rw, msgNotFound, http.StatusNotFound)
		return
	}

	if err = json.NewEncoder(rw).Encode(games); err != nil {
		log.Print(err)
		http.Error(rw, msgNotFound, http.StatusInternalServerError)
	}
}
