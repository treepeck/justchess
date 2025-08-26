// TODO: send email validation in signup.
// TODO: allow multiple sessions from different devices.
// TODO: automatically extend sessions without forcing players to sign in daily.
package auth

import (
	"database/sql"
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
func (s *AuthService) handleSignup(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, "Malformed request body", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !nameEx.MatchString(name) || !emailEx.MatchString(email) ||
		!pwdEx.MatchString(password) {
		http.Error(rw, "Unacceptable input", http.StatusNotAcceptable)
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(rw, "Cannot generate hash from password", http.StatusInternalServerError)
		return
	}

	dto := player.InsertPlayerDTO{
		Id:           randgen.GenBase62(),
		Name:         name,
		Email:        email,
		PasswordHash: string(pwdHash),
	}

	if err = player.Insert(s.pool, dto); err != nil {
		http.Error(rw, "Not unique name or email", http.StatusConflict)
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
func (s *AuthService) handleSignin(rw http.ResponseWriter, r *http.Request) {
	/*
		if err := r.ParseForm(); err != nil {
			http.Error(rw, "Malformed request body", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		if !emailEx.MatchString(email) || !pwdEx.MatchString(password) {
			http.Error(rw, "Unacceptable input", http.StatusNotAcceptable)
			return
		}

		p, err := db.SelectPlayerByEmail(email)
		if err != nil {
			http.Error(rw, "Invalid name or password", http.StatusNotAcceptable)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(password))
		if err != nil {
			http.Error(rw, "Invalid name or password", http.StatusNotAcceptable)
			return
		}

		sid := rand.Text()
		err = db.InsertSession(sid, p.Id)
		if err != nil {
			http.Error(rw, "Already signed", http.StatusConflict)
			return
		}

		setAutorizationCookie(rw, sid)
	*/
}

// CtxKey is used as a context type which provides player id.
type CtxKey string

const PidKey CtxKey = "pid"

/*
AuthorizeRequest authorizes the incoming request and passes a context containing
the player's credentials to the next handler function.  Sensitive endpoints must
check the subject's role.  It cannot be used to authenticate WebSocket handshake
request, since those do not support user-defined request headers.
*/
func AuthorizeRequest(next http.HandlerFunc) http.HandlerFunc {
	/*
		return func(rw http.ResponseWriter, r *http.Request) {
			if len(r.Cookies()) != 1 || r.Cookies()[0].Name != "Authorization" {
				http.Error(rw, "Unauthorized request.", http.StatusUnauthorized)
				return
			}

			sid := r.Cookies()[0].Value

			err := db.DeleteExpiredSessions()
			if err != nil {
				http.Error(rw, "Internal server error.", http.StatusInternalServerError)
				return
			}

			pid, err := db.SelectPlayerIdBySessionId(sid)
			if err != nil {
				http.Error(rw, "Session expired.", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), PidKey, pid)
			next.ServeHTTP(rw, r.WithContext(ctx))
		}
	*/
	return func(w http.ResponseWriter, r *http.Request) {}
}

func setAutorizationCookie(rw http.ResponseWriter, sid string) {
	c := http.Cookie{
		Name:     "Authorization",
		Value:    sid,
		Path:     "/",
		MaxAge:   86400, // Session will last 1 day.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &c)
}
