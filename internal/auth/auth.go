/*
Package auth implements authorization and authentication.
TODO: send email validation in signup.
TODO: allow multiple sessions from different devices.
TODO: automatically extend sessions without forcing players to sign in daily.
*/
package auth

import (
	"context"
	"database/sql"
	"io"
	"justchess/internal/player"
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
AuthService wraps a database connection pool and provides methods for handling
authorization and authentication HTTP requests.
*/
type AuthService struct {
	pool *sql.DB
}

/*
InitAuthService creates a new [AuthService], initializes the session table and
adds routes to the specified mux.
*/
func InitAuthService(pool *sql.DB, mux *http.ServeMux) error {
	s := AuthService{pool: pool}

	// Initializing session table.
	if _, err := pool.Exec(createQuery); err != nil {
		return err
	}

	mux.HandleFunc("POST /auth/signup", s.handleSignup)
	mux.HandleFunc("POST /auth/signin", s.handleSignin)
	mux.HandleFunc("POST /auth/verify", s.handleVerify)

	return nil
}

/*
handleSignup registers a new player.

The registration process includes the following steps:
 1. Decode the request body with the registration data.
 2. Validate the registration data using regular expressions.
 3. Hash the password to securely store it in the database.
 4. Insert a new player record.

The newly created user will not be authorized.
*/
func (s AuthService) handleSignup(rw http.ResponseWriter, r *http.Request) {
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

	dto := player.InsertPlayerDTO{
		Id:           randgen.GenId(randgen.IdLen),
		Name:         name,
		Email:        email,
		PasswordHash: string(pwdHash),
	}

	if err = player.Insert(s.pool, dto); err != nil {
		http.Error(rw, "Not unique name or email.", http.StatusConflict)
	}
}

/*
handleSignin authenticates a player by the provided credentials.

The authentication process includes the following steps:
 1. Decode the request body and extract the credentials.
 2. Validate the credentials using regular expressions.
 3. Retrieve the player data from the database using the email from request.
 4. Compare the stored password hash with the provided password.
 5. If the credetials are valid, verify that player isn't already authenticated.
 6. Create a new session.
 7. Respond with an authorization cookie and the player data.
*/
func (s AuthService) handleSignin(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, "Malformed request body", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if !emailEx.MatchString(email) || !pwdEx.MatchString(password) {
		http.Error(rw, "Malformed request body.", http.StatusNotAcceptable)
		return
	}

	p, err := player.SelectByEmail(s.pool, email)
	if err != nil {
		http.Error(rw, "Invalid credentials.", http.StatusNotAcceptable)
		return
	}

	if err = DeleteExpired(s.pool); err != nil {
		http.Error(rw, "Database cannot be accepted. Please, try again later.", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(password))
	if err != nil {
		http.Error(rw, "Invalid credentials.", http.StatusNotAcceptable)
		return
	}

	sessionId := randgen.GenId(randgen.SessionIdLen)
	err = Insert(s.pool, sessionId, p.Id)
	if err != nil {
		http.Error(rw, "Cannot create a new session. Please try again after 24 hours.", http.StatusConflict)
		return
	}

	setAutorizationCookie(rw, sessionId)
}

/*
handleVerify verifies that the provided in a request body session id is valid.
If it is, the player's data will be returned.  Request body must be a plain text
with a session id value.
*/
func (s AuthService) handleVerify(rw http.ResponseWriter, r *http.Request) {
	sessionId, err := io.ReadAll(r.Body)
	if err != nil || len(sessionId) != 32 {
		http.Error(rw, "Unauthorized request.", http.StatusBadRequest)
		return
	}

	if err := DeleteExpired(s.pool); err != nil {
		http.Error(rw, "Internal server error.", http.StatusInternalServerError)
		return
	}

	session, err := SelectById(s.pool, string(sessionId))
	if err != nil {
		http.Error(rw, "Session expired.", http.StatusUnauthorized)
		return
	}

	p, err := player.SelectById(s.pool, session.PlayerId)
	if err != nil {
		http.Error(rw, "Player was deleted.", http.StatusNotFound)
		return
	}

	rw.Header().Add("Content-Type", "plain/text")
	if _, err := rw.Write([]byte(p.Id)); err != nil {
		http.Error(rw, "Please try again later.", http.StatusInternalServerError)
	}
}

// CtxKey is used as a context type which provides player id.
type CtxKey string

const PidKey CtxKey = "pid"

/*
AuthorizeRequest authorizes the incoming request and passes a context containing
the player's credentials to the next handler function.
*/
func AuthorizeRequest(pool *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if len(r.Cookies()) != 1 || r.Cookies()[0].Name != "Authorization" {
			http.Error(rw, "Unauthorized request.", http.StatusUnauthorized)
			return
		}

		sessionId := r.Cookies()[0].Value

		if err := DeleteExpired(pool); err != nil {
			http.Error(rw, "Internal server error.", http.StatusInternalServerError)
			return
		}

		session, err := SelectById(pool, sessionId)
		if err != nil {
			http.Error(rw, "Session expired.", http.StatusUnauthorized)
			return
		}

		p, err := player.SelectById(pool, session.PlayerId)
		if err != nil {
			http.Error(rw, "Player was deleted.", http.StatusNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), PidKey, p.Id)
		next.ServeHTTP(rw, r.WithContext(ctx))
	}
}

func setAutorizationCookie(rw http.ResponseWriter, sessionId string) {
	c := http.Cookie{
		Name:     "Auth",
		Value:    sessionId,
		Path:     "/",
		MaxAge:   86400, // Session will last for 24 hours.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &c)
}
