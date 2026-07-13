// Package api implements HTTP API that serves database data.
package api

import (
	"encoding/json"
	"github.com/treepeck/chego"
	"justchess/internal/auth"
	"justchess/internal/db"
	"justchess/internal/randgen"
	"justchess/internal/response"
	"log"
	"math/rand/v2"
	"net/http"
	"time"
)

type Service struct {
	gameRepo   db.GameRepo
	playerRepo db.PlayerRepo
}

func NewService(gr db.GameRepo, pr db.PlayerRepo) Service {
	return Service{
		gameRepo:   gr,
		playerRepo: pr,
	}
}

func (s Service) RegisterRoutes(authService auth.Service, mux *http.ServeMux) {
	mux.HandleFunc("POST /api/engine", authService.MustAuthorize(s.createEngineGame))
	mux.HandleFunc("GET /api/engine-brief", s.engineBrief)
	mux.HandleFunc("GET /api/rated-brief", s.ratedBrief)
}

func (s Service) createEngineGame(rw http.ResponseWriter, r *http.Request) {
	session, ok := r.Context().Value(auth.SessionKey).(auth.Session)
	if !ok {
		log.Print("request context is broken")
		return
	}

	var c chego.Color
	if rand.IntN(2) == 1 {
		c = chego.ColorBlack
	}

	var d db.EngineDifficulty
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil ||
		d < db.Easy || d > db.Impossible {
		http.Error(rw, response.BadRequest, http.StatusBadRequest)
		return
	}
	gameId := randgen.GenId(randgen.IdLen)
	if session.IsGuest {
		session.Id = ""
	}
	if err := s.gameRepo.InsertEngineGame(gameId, session.Id, c, d); err != nil {
		log.Print(err)
		http.Error(rw, response.InternalError, http.StatusInternalServerError)
		return
	}
	http.Redirect(rw, r, "/engine/"+gameId, http.StatusFound)
}

func (s Service) engineBrief(rw http.ResponseWriter, r *http.Request) {
	// Mandatory parameter.
	playerId := r.URL.Query().Get("pid")
	if len(playerId) != 12 {
		http.Error(rw, response.BadRequest, http.StatusBadRequest)
		return
	}
	// Optional parameters for cursor-based pagination.
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cca"))

	var games []db.EngineBrief
	if err == nil && len(cursorId) == 12 {
		// If optional pagination parameters are defined.
		games, err = s.gameRepo.SelectEngineBrief(playerId, &db.Pagination{
			CursorId:        cursorId,
			CursorCreatedAt: cursorCreatedAt,
		})
	} else {
		games, err = s.gameRepo.SelectEngineBrief(playerId, nil)
	}

	if err != nil {
		http.Error(rw, response.NotFound, http.StatusNotFound)
		return
	}

	if err = json.NewEncoder(rw).Encode(games); err != nil {
		log.Print(err)
		http.Error(rw, response.InternalError, http.StatusInternalServerError)
		return
	}
	rw.Header().Add("Content-Type", "application/json")
}

func (s Service) ratedBrief(rw http.ResponseWriter, r *http.Request) {
	// Mandatory parameter.
	playerId := r.URL.Query().Get("pid")
	if len(playerId) != 12 {
		http.Error(rw, response.BadRequest, http.StatusBadRequest)
		return
	}
	// Optional parameters for cursor-based pagination.
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cca"))

	var games []db.RatedBrief
	if err == nil && len(cursorId) == 12 {
		// If optional pagination parameters are defined.
		games, err = s.gameRepo.SelectRatedBrief(playerId, &db.Pagination{
			CursorId:        cursorId,
			CursorCreatedAt: cursorCreatedAt,
		})
	} else {
		games, err = s.gameRepo.SelectRatedBrief(playerId, nil)
	}

	if err != nil {
		http.Error(rw, response.NotFound, http.StatusNotFound)
		return
	}

	if err = json.NewEncoder(rw).Encode(games); err != nil {
		log.Print(err)
		http.Error(rw, response.InternalError, http.StatusInternalServerError)
		return
	}
	rw.Header().Add("Content-Type", "application/json")
}
