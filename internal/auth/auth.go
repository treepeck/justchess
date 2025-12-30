// Package auth implements authorization and authentication.
// TODO: send email validation in signup.
package auth

import (
	"justchess/internal/db"
	"justchess/internal/randgen"
	"net/http"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// Regular expressions to validate user input.
var (
	nameEx  = regexp.MustCompile(`^[a-zA-Z0-9]{2,60}$`)
	emailEx = regexp.MustCompile(`^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$`)
	pwdEx   = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$`)
)

// Declaration of error messages.
const (
	msgUnauthorized  string = "Invalid credentials"
	msgBadRequest    string = "Malformed request body"
	msgConflict      string = "Not unique username or email"
	msgCannotHash    string = "Cannot generate password hash"
	msgDatabaseError string = "Database cannot be accessed. Please, try again later"
)

// Service wraps the database repository and provides methods for handling
// authorization and authentication HTTP requests.
type Service struct {
	repo db.Repo
}

func NewService(r db.Repo) Service { return Service{repo: r} }

// RegisterRoutes registers enpoints to the specified ServeMux.
func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/signup", s.signup)
	mux.HandleFunc("POST /auth/signin", s.signin)
}

// signup registers a new player.
//
// The registration process includes the following steps:
//  1. Decode the request body with the registration data.
//  2. Validate the registration data using regular expressions.
//  3. Hash the password to securely store it in the database.
//  4. Insert a new player record.
//  5. Creates a new session for the user.
func (s Service) signup(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, msgBadRequest, http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !nameEx.MatchString(name) || !emailEx.MatchString(email) ||
		!pwdEx.MatchString(password) {
		http.Error(rw, msgBadRequest, http.StatusNotAcceptable)
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(rw, msgCannotHash, http.StatusInternalServerError)
		return
	}

	playerId := randgen.GenId(randgen.IdLen)
	if s.repo.InsertPlayer(playerId, name, email, pwdHash) != nil {
		http.Error(rw, msgConflict, http.StatusConflict)
		return
	}

	s.genSession(rw, playerId)
}

// signin authenticates a player by the provided credentials.
//
// The authentication process includes the following steps:
//  1. Decode the request body and extract the credentials.
//  2. Validate the credentials using regular expressions.
//  3. Retrieve the player data from the database using the email from request.
//  4. Compare the stored password hash with the provided password.
//  5. Get all non-expired sessions with the same player_id.
//  6. If the number of sessions is more than or equal to five, remove the
//     oldest created session.
//  7. Create a new session.
//  8. Insert a newly created session.
//  9. Respond with an authorization cookie and the player data.
func (s Service) signin(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, msgBadRequest, http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if !emailEx.MatchString(email) || !pwdEx.MatchString(password) {
		http.Error(rw, msgBadRequest, http.StatusBadRequest)
		return
	}

	p, err := s.repo.SelectPlayerByEmail(email)
	if err != nil {
		http.Error(rw, msgUnauthorized, http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(p.PasswordHash, []byte(password))
	if err != nil {
		http.Error(rw, msgUnauthorized, http.StatusUnauthorized)
		return
	}

	sessions, err := s.repo.SelectSessionsByPlayerId(p.Id)
	if err != nil {
		http.Error(rw, msgDatabaseError, http.StatusInternalServerError)
		return
	}

	if len(sessions) == 5 {
		// Find the oldest session.
		min := sessions[0].CreatedAt
		ind := 0
		for i := 1; i < 5; i++ {
			if sessions[i].CreatedAt.Before(min) {
				min = sessions[i].CreatedAt
				ind = i
			}
		}

		// Delete the oldest session to replace it with the new one.
		if err = s.repo.DeleteSessionById(sessions[ind].Id); err != nil {
			http.Error(rw, msgDatabaseError, http.StatusInternalServerError)
			return
		}
	}

	s.genSession(rw, p.Id)
}

// genSession inserts a new record in the session table and adds the HTTP-only
// secure cookie to the response.
func (s Service) genSession(rw http.ResponseWriter, playerId string) {
	// Use generated unique string as session value.
	sessionId := randgen.GenId(randgen.SessionIdLen)

	if err := s.repo.InsertSession(sessionId, playerId); err != nil {
		http.Error(rw, msgDatabaseError, http.StatusInternalServerError)
		return
	}

	http.SetCookie(rw, &http.Cookie{
		Name:     "Auth",
		Value:    sessionId,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // Session will last for 30 days.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}
