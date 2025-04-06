// Package game implements game-related endpoints and
// provides access the the game db table.
package game

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /game/id/{id}", getById)
	mux.HandleFunc("GET /game/player/{id}", getByPlayerId)
	return mux
}

// getById returns detailed game info with decompressed moves.
func getById(rw http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	g, err := selectById(id.String())
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(g)
	if err != nil {
		log.Printf("cannot encode game: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

// getByPlayerId returns game info without detailed decompressed info.
// Used to display all played games in profile.
func getByPlayerId(rw http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	g, err := selectByPlayerId(id.String())
	if len(g) < 1 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(g)
	if err != nil {
		log.Printf("cannot encode games: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
	}
}
