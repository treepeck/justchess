// Package player implements player-data related endpoints and
// provides the access to the player-related tables.
package player

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /player/id/{id}", getById)
	mux.HandleFunc("GET /player/name/{name}", getByName)
	return mux
}

func getByName(rw http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if len(name) < 2 { // Minimal length of username.
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := SelectPlayerByName(name)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(p)
	if err != nil {
		log.Printf("cannot encode response: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getById(rw http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := SelectPlayerById(id.String())
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(p)
	if err != nil {
		log.Printf("cannot encode response: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
