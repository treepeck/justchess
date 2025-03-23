package user

import (
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           uuid.UUID `json:"id"`
	Name         string    `json:"username"`
	RegisteredAt time.Time `json:"registeredAt"`
	PasswordHash string    `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	Mail         string    `json:"-"`
	IsVerified   bool      `json:"-"`
}

type Register struct {
	Mail     string `json:"mail"`
	Name     string `json:"username"`
	Password string `json:"password"`
}

func CreateUserHandler(rw http.ResponseWriter, r *http.Request) {
	var req Register
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the provided data.
	if len(req.Name) < 1 || len(req.Name) > 36 ||
		len(req.Password) < 5 || len(req.Password) > 72 {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// Are name and mail unique?
	if !isUniqueNameAndMail(req.Name, req.Mail) {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	// Store the user data in a special table for unverified users.
	// Covert plain password to a hash to store in a db.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Password = string(hash)

	id, err := insertUnverified(req)
	if err != nil || len(id) == 0 {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	err = sendMailVerification(req.Mail, id)
	if err != nil {
		log.Printf("cannot send verification email: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
	}
}

func VerifyHandler(rw http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// Make sure that mail is not verified yet.
	if !isUnverifiedId(id.String()) {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	// Move data from the unverified to the main users table.
	u, err := deleteUnverified(id.String())
	if err != nil || u.Id == uuid.Nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}

	err = insertUser(u)
	if err != nil {
		rw.WriteHeader(http.StatusConflict)
		return
	}
}

func sendMailVerification(addr, id string) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("SMTP_MAIL"),
		os.Getenv("SMTP_PASSWORD"),
		"smtp.gmail.com",
	)

	msg := []byte("Subject: Welcome to Justchess!\r\n" +
		"\r\n" +
		"This email was sent due to a new account creation on justchess.org.\r\n" +
		"If it wasn't you, simply ignore this email.\r\n" +
		"\r\n" +
		"To confirm the registration, follow the link below:\r\n" +
		os.Getenv("DOMAIN") + "auth/verify?id=" + id + "\r\n")

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"support@justchess.org",
		[]string{addr},
		msg,
	)
}
