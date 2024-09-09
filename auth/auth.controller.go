package auth

import (
	"chess-api/models"
	"chess-api/repository"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
)

func handleGetUserByCookie(rw http.ResponseWriter, r *http.Request) {
	fn := slog.String("func", "handleGetUserByCookie")

	cookie, err := r.Cookie("UserId")
	if err != nil {
		slog.Warn("UserId cookie not found", fn, "err", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(cookie.Value)
	if err != nil {
		slog.Warn("userId cannot be Parsed", fn, "err", err)
		rw.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	u := repository.FindUserById(userId)
	if u == nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	// update user cookie to keep the user signed in
	setUserIdCookie(rw, u.Id)

	// send user back to the client
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(*u)
}

func handleGuest(rw http.ResponseWriter, r *http.Request) {
	fn := slog.String("func", "handleGuest")

	// decode the request body
	var cu models.CreateUserDTO
	err := json.NewDecoder(r.Body).Decode(&cu)
	if err != nil {
		rw.WriteHeader(http.StatusUnprocessableEntity)
		slog.Warn("error while decoding the request body ", fn, "err", err)
		return
	}

	// create a new user
	u := repository.AddGuest(cu.Id)
	if u == nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	// complete the guest authorization by setting a cookie with the user id
	setUserIdCookie(rw, u.Id)

	// send user back to the client
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(*u)
}

// When the guest is authorized, we set a http-only secure cookie with the
// player id to identify that the user authorized later.
func setUserIdCookie(rw http.ResponseWriter, userId uuid.UUID) {
	// set a http-only cookie with user id
	cookie := http.Cookie{
		Name:     "UserId",
		Value:    userId.String(),
		Path:     "/",
		Domain:   os.Getenv("SERVER_HOST"),
		MaxAge:   259200, // 3 days in seconds
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &cookie)
}
