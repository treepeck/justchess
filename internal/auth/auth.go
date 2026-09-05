// Package auth implements authorization and authentication.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

	"justchess/internal/db"
	"justchess/internal/randgen"

	"golang.org/x/crypto/bcrypt"
)

var (
	nameEx  = regexp.MustCompile(`^[a-zA-Z0-9]{2,60}$`)
	emailEx = regexp.MustCompile(`^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$`)
	pwdEx   = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$`)

	msgSignup = []byte("Please, check your email to confirm the registration. It may take several minutes for the email to be delivered and it may end up in spam.")
	msgReset  = []byte("Please, check your email to confirm the password reset. It may take several minutes for the email to be delivered and it may end up in spam.")
)

// Service wraps the database repositories and provides methods for handling
// authorization and authentication of HTTP requests.
type Service struct {
	// cookieKey is used to sign the secure Cookie.
	cookieKey []byte
	repo      db.AuthRepo
	// Store parsed emails to avoid expensive template parsing on each signup
	// or password reset.
	// First template is email-signup.tmpl.
	// Second template is email-reset.tmpl.
	emails [2]*template.Template
}

func NewService(key []byte, ar db.AuthRepo) Service { return Service{cookieKey: key, repo: ar} }

// RegisterRoutes registers enpoints to the specified ServeMux.
func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/signup", s.signup)
	mux.HandleFunc("POST /auth/signin", s.signin)
	mux.HandleFunc("POST /auth/reset-password", s.resetPassword)
	mux.HandleFunc("POST /auth/confirm-signup/{token}", s.confirmSignup)
	mux.HandleFunc("POST /auth/confirm-reset/{token}", s.confirmReset)
}

// signup registers a new player.
//
// The registration process includes the following steps:
//  1. Decode the request body with the registration data.
//  2. Validate the registration data using regular expressions.
//  3. Ensure that provided name and email are unique.
//  4. Store signup token in the database.
//  5. Send the verification email.
//
// If the verification email fails to send, the token insertion is rolled back,
// letting the player try again.
func (s Service) signup(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	name := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !nameEx.MatchString(name) || !emailEx.MatchString(email) ||
		!pwdEx.MatchString(password) {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	unique, err := s.repo.IsEmailUnique(email)
	if err != nil || !unique {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	token := randgen.GenId(randgen.SecureIdLen)
	if err = s.repo.InsertSignupToken(
		token,
		db.SignupData{
			Name: name, Email: email, PasswordHash: pwdHash,
		},
	); err != nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	url := os.Getenv("CONFIRM_SIGNUP_ENDPOINT") + token
	var buff bytes.Buffer
	if err = s.emails[0].Execute(&buff, tmplData{Name: name, Url: url}); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		panic("cannot execute email template")
	}

	body, err := json.Marshal(emailPayload{
		From:     sender{Email: os.Getenv("EMAIL_FROM")},
		To:       [1]reciever{{email}},
		Subject:  "Signup Verification",
		Category: "Transactional",
		Html:     buff.String(),
	})
	if err != nil {
		log.Printf("cannot encode email: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.sendEmail(body); err != nil {
		log.Printf("cannot send email: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		// Remove inserted token.
		if err = s.repo.DeleteSignupToken(token); err != nil {
			log.Printf("%v cannot delete sign up token %s\n", err, token)
		}
		return
	}
	// Write information message after successfull sign up.
	rw.Write(msgSignup)
}

// signin authenticates a player by the provided credentials.
//
// The authentication process includes the following steps:
//  1. Decode the request body and extract the credentials.
//  2. Validate the credentials using regular expressions.
//  3. Retrieve the player data from the database using the email from request.
//  4. Compare the stored password hash with the provided password.
//  5. Respond with [secureCookie] and the player data.
func (s Service) signin(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if !emailEx.MatchString(email) || !pwdEx.MatchString(password) {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := s.repo.SelectCredentialsByEmail(email)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(c.PasswordHash, []byte(password))
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	s.setSecureCookie(rw, c.Id, false)
	// Redirect to the home page after successfull sign in.
	http.Redirect(rw, r, "/", http.StatusFound)
}

// If the verification email fails to send, the token insertion is rolled back,
// letting the player try again.
func (s Service) resetPassword(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if !emailEx.MatchString(email) || !pwdEx.MatchString(password) {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := s.repo.SelectIdentityByEmail(email)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("%v cannot generate hash from password: %s\n", err, password)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	token := randgen.GenId(randgen.SecureIdLen)
	if err = s.repo.InsertPasswordResetToken(token, p.Id, pwdHash); err != nil {
		log.Print(err)
		rw.WriteHeader(http.StatusConflict)
		return
	}

	url := os.Getenv("CONFIRM_RESET_ENDPOINT") + token
	var buff bytes.Buffer
	if err = s.emails[1].Execute(&buff, tmplData{Name: p.Name, Url: url}); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		panic("cannot execute email template")
	}

	body, err := json.Marshal(emailPayload{
		From:     sender{Email: os.Getenv("EMAIL_FROM")},
		To:       [1]reciever{{email}},
		Subject:  "Password Reset",
		Category: "Transactional",
		Html:     buff.String(),
	})
	if err != nil {
		log.Printf("cannot encode email: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.sendEmail(body); err != nil {
		log.Printf("cannot send email: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		// Remove inserted token.
		if err = s.repo.DeletePasswordResetToken(token); err != nil {
			log.Printf("%v cannot delete password reset token %s\n", err, token)
		}
		return
	}
	// Write information message after successfull password reset.
	rw.Write(msgReset)
}

// confirmSignup completes the registration process for players who click the
// confirmation email link.
//
// The confirmation process includes the following steps:
//  1. Fetch signup credentials from database using provided token.
//  2. Insert new player record using provided credentials.
//  3. Delete used token.
//  4. Generate [secureCookie] for the player.
func (s Service) confirmSignup(rw http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	data, err := s.repo.SelectSignupDataByToken(token)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	id := randgen.GenId(randgen.IdLen)
	if err = s.repo.InsertPlayer(id, data); err != nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}
	if err = s.repo.DeleteSignupToken(token); err != nil {
		log.Printf("%v cannot delete sign up token %s\n", err, token)
	}
	s.setSecureCookie(rw, id, false)
	// Redirect user to home page after successfull signup confirmation.
	http.Redirect(rw, r, "/", http.StatusFound)
}

// confirmReset completes the password reset process by updating the player
// password and deleting the used password_reset_token.
func (s Service) confirmReset(rw http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	c, err := s.repo.SelectCredentialsByResetToken(token)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if err = s.repo.UpdatePasswordHash(c.Id, c.PasswordHash); err != nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	if err = s.repo.DeletePasswordResetToken(token); err != nil {
		log.Printf("%v cannot delete password reset token %s\n", err, token)
	}
	// Redirect user to sign in page after successfull password reset.
	http.Redirect(rw, r, "/signin", http.StatusFound)
}

type contextKey int

const SessionKey contextKey = 0

// Authorize is a middleware that validates the Auth Cookie. If cookie
// is expired, missing, or not valid, the new Guest cookie is generated.
// NOTE: This middleware should be applied to general endpoints that
// are expected to handle the guest traffic.
func (s Service) Authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		session := Session{Id: randgen.GenId(randgen.IdLen), IsGuest: true}
		c, err := r.Cookie("Auth")
		if err == nil {
			session, err = validateSecureCookie([]byte(c.Value), s.cookieKey)
			if err != nil {
				session.Id = randgen.GenId(randgen.IdLen)
			}
		}
		ctx := context.WithValue(r.Context(), SessionKey, session)
		next.ServeHTTP(rw, r.WithContext(ctx))
	}
}

// MustAuthorize is a middleware that validates the Auth Cookie. If cookie
// is expired, missing, or not valid, the http.StatusUnauthorized is
// written to response.
// NOTE: This middleware should be applied to strictly protected endpoints
// which are not expected to handle the guest traffic.
func (s Service) MustAuthorize(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("Auth")
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		session, err := validateSecureCookie([]byte(c.Value), s.cookieKey)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), SessionKey, session)
		next.ServeHTTP(rw, r.WithContext(ctx))
	}
}

func (s Service) setSecureCookie(rw http.ResponseWriter, id string, isGuest bool) {
	val, err := genSecureCookie(Session{Id: id, IsGuest: isGuest}, s.cookieKey)
	if err != nil {
		log.Println("cannot generate secure cookie: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Name:     "Auth",
		Value:    val,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
