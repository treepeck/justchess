// Package auth implements player authentication and authorization.
// Sign up and Sign in are both considered as authentication.
package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"justchess/pkg/db"
	"justchess/pkg/player"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type signIn struct {
	// Login represents username or mail.
	Login    string `json:"login"`
	Password string `json:"password"`
}

type mailData struct {
	Name            string
	VerificationURL string
}

type passwordReset struct {
	Mail     string `json:"mail"`
	Password string `json:"password"` // New password.
}

type PlayerDTO struct {
	Player      player.Player `json:"player"`
	Role        Role          `json:"role"`
	AccessToken string        `json:"accessToken"`
}

// Regular expressions for validating registration data.
var (
	nameRE = regexp.MustCompile(`[a-zA-Z]{1}[a-zA-Z0-9_]+`)
	mailRE = regexp.MustCompile(`[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+`)
)

func Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /auth/refresh", refreshHandler)
	mux.HandleFunc("GET /auth/guest", guestHandler)
	mux.HandleFunc("GET /auth/verify", verifyHandler)
	mux.HandleFunc("POST /auth/signup", signUpHandler)
	mux.HandleFunc("POST /auth/signin", signInHandler)
	mux.HandleFunc("POST /auth/reset", passwordResetHandler)
	return mux
}

///////////////////////////////////////////////////////////////
//                       AUTHENTICATION                      //
///////////////////////////////////////////////////////////////

// signUpHandler rollbacks the insert if the confirmation mail cannot be sent.
func signUpHandler(rw http.ResponseWriter, r *http.Request) {
	var req player.Register
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the provided data.
	if !nameRE.Match([]byte(req.Name)) || !mailRE.Match([]byte(req.Mail)) ||
		len(req.Password) < 5 || len(req.Password) > 72 {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// Are the name and mail unique?
	if player.IsTakenNameOrMail(req.Name, req.Mail) {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	// Covert plain password to a hash to store in a db.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("cannot generate password hash: %v\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Password = string(hash)

	// Begin a new transaction to be able to roll-back the insert in case the email sending will go wrong.
	tx, err := db.Pool.Begin()
	if err != nil {
		log.Printf("Cannot begin a new transaction: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Generate unique token to safely confirm the registration.
	token := rand.Text()
	err = player.InsertTokenRegistration(token, req, tx)
	defer tx.Rollback()
	if err != nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	data := mailData{
		Name:            req.Name,
		VerificationURL: os.Getenv("DOMAIN") + "/verify?action=registration&token=" + token,
	}

	err = sendMail(req.Mail, "Subject: Email Verification\r\n",
		"./templates/mail-verify.html", data)
	if err != nil {
		log.Printf("cannot send verification email: %v\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("cannot commit transaction: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

// verifyHandler verifies email registration and password resets.
// Info about email registrations is stored in the 'unverified' table.
// Info about pending password resets is stored in the 'tokenreset' table.
func verifyHandler(rw http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	token := r.URL.Query().Get("token")
	if token == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	switch action {
	case "reset":
		completeReset(rw, token)

	case "registration":
		completeSignUp(rw, token)

	default:
		rw.WriteHeader(http.StatusBadRequest)
	}
}

func signInHandler(rw http.ResponseWriter, r *http.Request) {
	var req signIn
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := player.SelectPlayerByLogin(req.Login)
	// TODO: add bruteforce protection.
	if err != nil ||
		bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(req.Password)) != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	completeAuth(rw, p, RolePlayer)
}

func passwordResetHandler(rw http.ResponseWriter, r *http.Request) {
	var req passwordReset
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.Password) < 5 || len(req.Password) > 72 {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := player.SelectPlayerByMail(req.Mail)
	if err != nil || p.Id == uuid.Nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	token := rand.Text()
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("cannot generate hash: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := db.Pool.Begin()
	if err != nil {
		log.Printf("cannot begin transaction: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	err = player.InsertTokenReset(token, p.Id.String(), string(hash), tx)
	if err != nil {
		log.Printf("cannot insert token reset: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := mailData{
		Name:            p.Name,
		VerificationURL: os.Getenv("DOMAIN") + "/verify?action=reset&token=" + token,
	}

	if err = sendMail(p.Mail, "Subject: Password Reset\r\n",
		"./templates/password-reset.html", data); err != nil {
		log.Printf("cannot send mail: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func completeSignUp(rw http.ResponseWriter, token string) {
	r, err := player.SelectTokenRegistration(token)
	if err != nil || r.Mail == "" {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	id := uuid.New()
	err = player.InsertPlayer(id.String(), r)
	if err != nil {
		log.Printf("cannot insert player %s: %v\n", r.Mail, err)
		rw.WriteHeader(http.StatusConflict)
		return
	}

	completeAuth(rw, player.Player{Id: id, Name: r.Name, CreatedAt: time.Now()}, RolePlayer)
}

// completeReset authenticates the player after password reset.
func completeReset(rw http.ResponseWriter, token string) {
	id, hash, err := player.SelectTokenReset(token)
	if err != nil || hash == "" {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	p, err := player.UpdatePasswordHash(hash, id)
	if err != nil || p.Id == uuid.Nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	completeAuth(rw, p, RolePlayer)
}

// sendMail assumes that dev.env file contains such variables:
//
//  1. mail of the sender (SMTP_MAIL);
//  2. password for that email (SMTP_PASSWORD).
func sendMail(addr, subject, templatePath string, data mailData) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("SMTP_MAIL"),
		os.Getenv("SMTP_PASSWORD"),
		"smtp.gmail.com",
	)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body, err := genTemplate(
		templatePath,
		data,
	)
	if err != nil {
		return err
	}

	msg := []byte(subject + mime + body)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"support@justchess.org",
		[]string{addr},
		msg,
	)
}

// genTemplate generates the mail html templates.
func genTemplate(path string, data any) (string, error) {
	t, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	err = t.Execute(&buff, data)
	return buff.String(), err
}

///////////////////////////////////////////////////////////////
//                       AUTHORIZATION                       //
///////////////////////////////////////////////////////////////

// refreshHandler is used when the access token becomes invalid.
func refreshHandler(rw http.ResponseWriter, r *http.Request) {
	encoded, err := parseRefreshTokenCookie(r)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	cms, err := DecodeToken(encoded, "REFRESH_TOKEN_SECRET")
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Select updated player info since the user_name may have been changed.
	p, err := player.SelectPlayerById(cms.Id.String())
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	completeAuth(rw, p, RolePlayer)
}

// guestHandler creates a guest player and sends back guest JWT.
func guestHandler(rw http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	guest := player.Player{
		Id:        id,
		Name:      "Guest-" + id.String()[0:8],
		CreatedAt: time.Now(),
	}
	completeAuth(rw, guest, RoleGuest)
}

// completeAuth generates the JWT pair and sends the refresh token
// as a HTTP-Only Secure Cookie and access token as a plain/text response body.
func completeAuth(rw http.ResponseWriter, p player.Player, r Role) {
	at, rt, err := generatePair(p.Id, p.Name, r)
	if err != nil {
		log.Printf("cannot generate pair: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(rw, rt)
	// Send access token as a response.
	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(PlayerDTO{Player: p, AccessToken: at, Role: r})
	if err != nil {
		log.Printf("cannot decode response: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func setRefreshTokenCookie(rw http.ResponseWriter, et string) {
	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer " + et,
		Path:     "/",
		MaxAge:   2592000, // Token will last 30 days.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &cookie)
}

func parseRefreshTokenCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		return "", err
	}
	if len(cookie.Value) < 100 || cookie.Value[:7] != "Bearer " {
		return "", err
	}
	return cookie.Value[7:], nil
}
