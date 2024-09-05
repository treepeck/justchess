package auth

import (
	"chess-api/models"
	"chess-api/repository"
	"encoding/json"
	"log"
	"net/http"
)

func handleGuest(rw http.ResponseWriter, r *http.Request) {
	// decode the request body
	var cu models.CreateUserDTO
	err := json.NewDecoder(r.Body).Decode(&cu)
	if err != nil {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		log.Println("HandleGuest: error while decoding the request body ", err)
		return
	}

	// create a new user
	u := repository.AddGuest(cu.Id)
	if u == nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	// send user back to the client
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(*u)
}
