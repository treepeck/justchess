// Package api implements HTTP API that serves database data.
package api

import (
	"encoding/json"
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
	mux.HandleFunc("POST /api/create-engine", authService.MustAuthorize(s.createEngineGame))
	mux.HandleFunc("GET /api/engine", s.engine)
	mux.HandleFunc("GET /api/rated", s.rated)
}

func (s Service) createEngineGame(rw http.ResponseWriter, r *http.Request) {
	// TODO: support time control and time bonus.
	session, ok := r.Context().Value(auth.SessionKey).(auth.Session)
	if !ok {
		log.Print("request context is broken")
		return
	}
	var d db.EngineDifficulty
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil ||
		d < db.Easy || d > db.Impossible {
		http.Error(rw, response.BadRequest, http.StatusBadRequest)
		return
	}

	stockfishId := db.EasyStockfishId
	switch d {
	case db.Medium:
		stockfishId = db.MediumStockfishId
	case db.Hard:
		stockfishId = db.HardStockfishId
	case db.Insane:
		stockfishId = db.InsaneStockfishId
	case db.Impossible:
		stockfishId = db.ImpossibleStockfishId
	}
	// 0 - White; 1 - Black.
	var whiteId, blackId string
	if rand.IntN(2) == 1 {
		if !session.IsGuest {
			whiteId = session.Id
		}
		blackId = stockfishId
	} else {
		if !session.IsGuest {
			blackId = session.Id
		}
		whiteId = stockfishId
	}

	gameId := randgen.GenId(randgen.IdLen)
	if err := s.gameRepo.Insert(db.Game{
		Id: gameId,
		White: db.Player{
			Id: whiteId,
		},
		Black: db.Player{
			Id: blackId,
		},
	}); err != nil {
		log.Print(err)
		http.Error(rw, response.InternalError, http.StatusInternalServerError)
		return
	}
	http.Redirect(rw, r, "/engine/"+gameId, http.StatusFound)
}

func (s Service) engine(rw http.ResponseWriter, r *http.Request) {
	// Mandatory parameter.
	id := r.URL.Query().Get("pid")
	if len(id) != 12 {
		http.Error(rw, response.BadRequest, http.StatusBadRequest)
		return
	}
	// Optional parameters for cursor-based pagination.
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cca"))

	var games []db.Game
	if err == nil && len(cursorId) == 12 {
		// If optional pagination parameters are defined.
		games, err = s.gameRepo.SelectByPlayerId(id, &db.Pagination{
			CursorId:        cursorId,
			CursorCreatedAt: cursorCreatedAt,
		})
	} else {
		games, err = s.gameRepo.SelectByPlayerId(id, nil)
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

func (s Service) rated(rw http.ResponseWriter, r *http.Request) {
	// Mandatory parameter.
	id := r.URL.Query().Get("pid")
	if len(id) != 12 {
		http.Error(rw, response.BadRequest, http.StatusBadRequest)
		return
	}
	// Optional parameters for cursor-based pagination.
	cursorId := r.URL.Query().Get("cid")
	cursorCreatedAt, err := time.Parse(time.RFC3339, r.URL.Query().Get("cca"))

	var games []db.Game
	if err == nil && len(cursorId) == 12 {
		// If optional pagination parameters are defined.
		games, err = s.gameRepo.SelectByPlayerId(id, &db.Pagination{
			CursorId:        cursorId,
			CursorCreatedAt: cursorCreatedAt,
		})
	} else {
		games, err = s.gameRepo.SelectByPlayerId(id, nil)
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
