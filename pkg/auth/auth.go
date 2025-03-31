// Package auth implements user authentication and authorization.
// Sign up and Sign in are both considered as authentication.
package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"justchess/pkg/db"
	"justchess/pkg/user"
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

type SignIn struct {
	// Login represents username or mail.
	Login    string `json:"login"`
	Password string `json:"password"`
}

type MailData struct {
	Name            string
	VerificationURL string
}

type PasswordReset struct {
	Mail     string `json:"mail"`
	Password string `json:"password"` // New password.
}

type UserDTO struct {
	User        user.User `json:"user"`
	AccessToken string    `json:"accessToken"`
	Role        Role      `json:"role"`
}

// Regular expressions for validating registration data.
var (
	nameRE = regexp.MustCompile(`[a-zA-Z]{1}[a-zA-Z0-9_]+`)
	mailRE = regexp.MustCompile(`[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+`)
)

///////////////////////////////////////////////////////////////
//                       AUTHENTICATION                      //
///////////////////////////////////////////////////////////////

// SignUpHandler validates that the provided mail and name are unique,
// stores the profile data in the 'unverified' DB table and sends the
// confirmation mail.
// If the confirmation mail cannot be sent, the transaction will be rolled-back.
func SignUpHandler(rw http.ResponseWriter, r *http.Request) {
	var req user.Register
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
	if user.IsTakenUsernameOrMail(req.Name, req.Mail) {
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

	id, err := user.InsertUnverified(req, tx)
	defer tx.Rollback()
	if err != nil || len(id) == 0 {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	data := MailData{
		Name:            req.Name,
		VerificationURL: os.Getenv("DOMAIN") + "verify/" + id,
	}

	err = sendMail(req.Mail, "Subject: Email Verification\r\n",
		"./templates/mail-verify.html", data)
	if err != nil {
		log.Printf("cannot send verification email: %v\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = tx.Commit(); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// VerifyMailHandler ensures that the mail is not verified yet.
// Is it doesn't, the profile data is moved from the 'unverified' table to the
// 'users' table and a pair of JWTs are generated for a newly created user.
func VerifyMailHandler(rw http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	idStr := id.String()

	tx, err := db.Pool.Begin()
	if err != nil {
		log.Printf("cannot begin a new transaction: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	data, err := user.DeleteUnverified(idStr, tx)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	u, err := user.InsertUser(idStr, data, tx)
	if err != nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	if err = tx.Commit(); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	completeAuth(rw, u, RoleUser)
}

func SignInHandler(rw http.ResponseWriter, r *http.Request) {
	var req SignIn
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := user.SelectByLogin(req.Login)
	// TODO: add bruteforce protection.
	if err != nil || u.Id == uuid.Nil ||
		bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	completeAuth(rw, u, RoleUser)
}

func PasswordResetHandler(rw http.ResponseWriter, r *http.Request) {
	var req PasswordReset
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.Password) < 5 || len(req.Password) > 72 {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := user.SelectByMail(req.Mail)
	if err != nil || u.Id == uuid.Nil {
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

	err = user.InsertTokenReset(token, u.Id.String(), string(hash), tx)
	if err != nil {
		log.Printf("cannot insert token reset: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := MailData{
		Name:            u.Name,
		VerificationURL: os.Getenv("DOMAIN") + "reset/" + token,
	}

	if err = sendMail(u.Mail, "Subject: Password Reset\r\n",
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

// sendMail assumes that dev.env file contains such variables:
//
//  1. mail of the sender (SMTP_MAIL);
//  2. password for that email (SMTP_PASSWORD).
func sendMail(addr, subject, templatePath string, data MailData) error {
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

// RefreshHandler is used when the access token becomes invalid.
func RefreshHandler(rw http.ResponseWriter, r *http.Request) {
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

	// Select updated user info since the user_name may have been changed.
	u, err := user.SelectById(cms.Id.String())
	if err != nil || u.Id == uuid.Nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	completeAuth(rw, u, RoleUser)
}

// GuestHandler creates a guest user and sends back guest JWT.
func GuestHandler(rw http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	guest := user.User{
		Id:           id,
		Name:         "Guest-" + id.String()[0:8],
		RegisteredAt: time.Now(),
	}
	completeAuth(rw, guest, RoleGuest)
}

// completeAuth generates the JWT pair and sends the refresh token
// as a HTTP-Only Secure Cookie and access token as a plain/text response body.
func completeAuth(rw http.ResponseWriter, u user.User, r Role) {
	at, rt, err := generatePair(u.Id, u.Name, r)
	if err != nil {
		log.Printf("cannot generate pair: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(rw, rt)
	// Send access token as a response.
	rw.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(UserDTO{User: u, AccessToken: at, Role: r})
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
