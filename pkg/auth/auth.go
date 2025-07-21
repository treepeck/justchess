// TODO: send email validation in signup.
// TODO: allow multiple sessions from different devices.
// TODO: automatically extend sessions without forcing players to sign in daily.
package auth

import (
	"crypto/rand"
	"encoding/json"
	"justchess/pkg/db"
	"log"
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

func Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/signup", signup)
	mux.HandleFunc("POST /auth/signin", signin)
	return mux
}

// signup registers a new player.
//
// The registration process includes the following steps:
//  1. Decode the request body with the registration data.
//  2. Validate the registration data using regular expressions.
//  3. Hash the password to securely store it in the database.
//  4. Insert a new player record.
//
// The newly created user will not be authorized.
func signup(rw http.ResponseWriter, r *http.Request) {
	var dto signupDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(rw, "Malformed request body", http.StatusBadRequest)
		return
	}

	if !nameEx.MatchString(dto.Name) || !emailEx.MatchString(dto.Email) ||
		!pwdEx.MatchString(dto.Password) {
		http.Error(rw, "Unacceptable input", http.StatusNotAcceptable)
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: not able to generate hash %v", err)
		http.Error(rw, "Cannot hash password", http.StatusInternalServerError)
		return
	}

	if err = db.InsertPlayer(dto.Name, dto.Email, string(pwdHash)); err != nil {
		http.Error(rw, "Not unique name or email", http.StatusConflict)
	}
}

// signin authenticates a player by the provided credentials.
//
// The authentication process includes the following steps:
//  1. Decode the request body and extract the credentials.
//  2. Validate the credentials using regular expressions.
//  3. Retrieve the player data from the database using the email from request.
//  4. Compare the stored password hash with the provided password.
//  5. If the credetials are valid, verify that player isn't already authenticated.
//  6. Create a new session.
//  7. Respond with an authorization cookie and the player data.
func signin(rw http.ResponseWriter, r *http.Request) {
	var dto signinDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(rw, "Malformed request body", http.StatusBadRequest)
		return
	}

	if !emailEx.MatchString(dto.Email) || !pwdEx.MatchString(dto.Password) {
		http.Error(rw, "Unacceptable input", http.StatusNotAcceptable)
		return
	}

	p, err := db.SelectPlayerByEmail(dto.Email)
	if err != nil {
		http.Error(rw, "Player not found", http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(dto.Password))
	if err != nil {
		http.Error(rw, "Invalid password", http.StatusNotAcceptable)
		return
	}

	sid := rand.Text()
	err = db.InsertSession(sid, p.Id)
	if err != nil {
		http.Error(rw, "Already signed", http.StatusConflict)
		return
	}

	setAutorizationCookie(rw, sid)
}
