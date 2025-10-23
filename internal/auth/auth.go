/*
Package auth implements authorization and authentication.
TODO: send email validation in signup.
TODO: allow multiple sessions from different devices.
TODO: automatically extend sessions without forcing players to sign in daily.
*/
package auth

import (
	"encoding/json"
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

/*
Service wraps the database repository and provides methods for handling
authorization and authentication HTTP requests.
*/
type Service struct {
	repo *db.Repo
}

func NewService(r *db.Repo) Service {
	return Service{repo: r}
}

/*
HandleSignup registers a new player.

The registration process includes the following steps:
 1. Decode the request body with the registration data.
 2. Validate the registration data using regular expressions.
 3. Hash the password to securely store it in the database.
 4. Insert a new player record.
 5. Creates a new session for the user.
*/
func (s Service) HandleSignup(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, "Malformed request body.", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !nameEx.MatchString(name) || !emailEx.MatchString(email) ||
		!pwdEx.MatchString(password) {
		http.Error(rw, "Malformed request body.", http.StatusNotAcceptable)
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(rw, "Cannot generate hash from password.", http.StatusInternalServerError)
		return
	}

	playerId := randgen.GenId(randgen.IdLen)
	if s.repo.InsertPlayer(playerId, name, email, pwdHash) != nil {
		http.Error(rw, "Not unique name or email.", http.StatusConflict)
		return
	}

	s.genSession(rw, playerId)
}

/*
HandleSignin authenticates a player by the provided credentials.

The authentication process includes the following steps:
 1. Decode the request body and extract the credentials.
 2. Validate the credentials using regular expressions.
 3. Retrieve the player data from the database using the email from request.
 4. Compare the stored password hash with the provided password.
 5. If the credetials are valid, verify that player isn't already authenticated.
 6. Create a new session.
 7. Respond with an authorization cookie and the player data.
*/
func (s Service) HandleSignin(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, "Malformed request body", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if !emailEx.MatchString(email) || !pwdEx.MatchString(password) {
		http.Error(rw, "Malformed request body.", http.StatusBadRequest)
		return
	}

	p, err := s.repo.SelectPlayerByEmail(email)
	if err != nil {
		http.Error(rw, "Invalid credentials.", http.StatusNotAcceptable)
		return
	}

	if err = s.repo.DeleteExpiredSessions(); err != nil {
		http.Error(rw, "Database cannot be accepted. Please, try again later.", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(p.PasswordHash, []byte(password))
	if err != nil {
		http.Error(rw, "Invalid credentials.", http.StatusNotAcceptable)
		return
	}

	s.genSession(rw, p.Id)
}

/*
CtxKey is used as a context type which provides player id.
*/
type CtxKey string

const PidKey CtxKey = "pid"

/*
HandleVerify validates the session ID extracted from the Authorization cookie
and, if valid, returns the player's data in the response.
*/
func (s Service) HandleVerify(rw http.ResponseWriter, r *http.Request) {
	if r.Cookies()[0].Name != "Authorization" {
		http.Error(rw, "Unauthorized request.", http.StatusUnauthorized)
		return
	}

	sessionId := r.Cookies()[0].Value

	if err := s.repo.DeleteExpiredSessions(); err != nil {
		http.Error(rw, "Internal server error.", http.StatusInternalServerError)
		return
	}

	p, err := s.repo.SelectPlayerBySessionId(sessionId)
	if err != nil {
		http.Error(rw, "Session not found.", http.StatusUnauthorized)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	if err = json.NewEncoder(rw).Encode(p); err != nil {
		http.Error(rw, "Please try again later.", http.StatusInternalServerError)
	}
}

/*
genSession inserts a new record in the session table and adds the HTTP-only
secure cookie to the response.
*/
func (s *Service) genSession(rw http.ResponseWriter, playerId string) {
	sessionId := randgen.GenId(randgen.SessionIdLen)
	if s.repo.InsertSession(sessionId, playerId) != nil {
		http.Error(rw, "Cannot create a new session. Please try again after 24 hours.", http.StatusConflict)
		return
	}

	c := http.Cookie{
		Name:     "Authorization",
		Value:    sessionId,
		Path:     "/",
		MaxAge:   86400, // Session will last for 24 hours.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &c)
}
