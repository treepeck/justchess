// Package player implements player-data related endpoints and
// provides the access to the player-related tables.
//
// All insert and delete operations are made using Transactions.
// It is a caller responsibility to end a transaction.
package player

import (
	"encoding/json"
	"log"
	"net/http"
)

func Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /player/{name}", getByNameHandler)
	// mux.HandleFunc("POST /player/comment", commentHandler)
	return mux
}

func getByNameHandler(rw http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if len(name) < 2 { // Minimal length of username.
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := SelectPlayerByName(name)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(u)
	if err != nil {
		log.Printf("cannot encode response: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
