package user

import (
	"chess-api/repository"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func handleGetUserById(rw http.ResponseWriter, r *http.Request) {
	// Extract user id from the URL path
	idStr := r.URL.Path[len("/id/"):]

	userId, err := uuid.Parse(idStr)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	user := repository.FindUserById(userId)
	if user == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(user)
}
