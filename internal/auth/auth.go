// Package auth implements authorization and authentication.
package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"

	"justchess/internal/db"
	"justchess/internal/randgen"

	"golang.org/x/crypto/bcrypt"
)

// Regular expressions to validate user input.
var (
	nameEx  = regexp.MustCompile(`^[a-zA-Z0-9]{2,60}$`)
	emailEx = regexp.MustCompile(`^[a-zA-Z0-9._]+@[a-zA-Z0-9._]+\.[a-zA-Z0-9._]+$`)
	pwdEx   = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+-/.<>]{5,71}$`)
)

const (
	// Declaration of error messages.
	msgUnauthorized    string = "Invalid credentials"
	msgBadRequest      string = "Malformed request body"
	msgConflict        string = "Not unique username or email"
	msgTokenConflict   string = "You already have a pending token"
	msgCannotHash      string = "Cannot generate password hash"
	msgCannotSendEmail string = "Cannot send email. Please, ensure that email is valid"
	msgDatabaseError   string = "Database cannot be accessed. Please, try again later"

	sessionsThreshold int = 5
)

// tmplData is a data object used to fill up the verification email while
// executing a template file.
type tmplData struct {
	Name string
	Url  string
}

type emailSender struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type emailReciever struct {
	Email string `json:"email"`
}

// deliveryServicePayload is needed to send email through a service such as Mailtrap.
type deliveryServicePayload struct {
	From     emailSender      `json:"from"`
	To       [1]emailReciever `json:"to"`
	Subject  string           `json:"subject"`
	Html     string           `json:"html"`
	Category string           `json:"category"`
}

// Service wraps the database repositories and provides methods for handling
// authorization and authentication of HTTP requests.
type Service struct {
	repo db.AuthRepo
	// Store parsed emails to avoid expensive template parsing on each signup
	// or password reset.
	// First template is signup_verification_email.tmpl.
	// Seconds template is password_reset_email.tmpl.
	emails [2]*template.Template
}

// InitService parses email templates and stores them in the [Service].
func InitService(repo db.AuthRepo, tmplPath string) (Service, error) {
	var emails [2]*template.Template

	signupTmpl, err := template.ParseFiles(tmplPath + "email_verify_signup.tmpl")
	if err != nil {
		return Service{}, err
	}
	emails[0] = signupTmpl

	resetTmpl, err := template.ParseFiles(tmplPath + "email_reset_password.tmpl")
	if err != nil {
		return Service{}, err
	}
	emails[1] = resetTmpl

	return Service{repo: repo, emails: emails}, err
}

// RegisterRoutes registers enpoints to the specified ServeMux.
func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/signup", s.signup)
	mux.HandleFunc("POST /auth/signin", s.signin)
	mux.HandleFunc("POST /auth/reset-password", s.resetPassword)
	mux.HandleFunc("/auth/verify-signup/{token}", s.verifySignup)
	mux.HandleFunc("/auth/verify-reset-password/{token}", s.verifyResetPassword)
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

	unique, err := s.repo.AreNameAndEmailUnique(name, email)
	if err != nil || !unique {
		http.Error(rw, msgConflict, http.StatusConflict)
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(rw, msgCannotHash, http.StatusInternalServerError)
		return
	}

	token := randgen.GenId(randgen.SecureIdLen)
	if err = s.repo.InsertSignupToken(
		token,
		db.SignupData{
			Name: name, Email: email, PasswordHash: pwdHash,
		},
	); err != nil {
		http.Error(rw, msgConflict, http.StatusConflict)
		return
	}

	url := os.Getenv("SIGNUP_VERIFY_ENDPOINT") + token
	var buff bytes.Buffer
	if err = s.emails[0].Execute(&buff, tmplData{Name: name, Url: url}); err != nil {
		log.Print(err)
		http.Error(rw, msgCannotSendEmail, http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(deliveryServicePayload{
		From:     emailSender{Email: os.Getenv("EMAIL_FROM")},
		To:       [1]emailReciever{{email}},
		Subject:  "Signup Verification",
		Category: "Transactional",
		Html:     buff.String(),
	})
	if err != nil {
		log.Print(err)
		http.Error(rw, msgCannotSendEmail, http.StatusInternalServerError)
		return
	}

	if err = s.sendEmail(body); err != nil {
		log.Print(err)
		http.Error(rw, msgCannotSendEmail, http.StatusInternalServerError)
		// Remove inserted token.
		if err = s.repo.DeleteSignupToken(token); err != nil {
			log.Print(err)
		}
	}
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

	c, err := s.repo.SelectCredentialsByEmail(email)
	if err != nil {
		http.Error(rw, msgUnauthorized, http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(c.PasswordHash, []byte(password))
	if err != nil {
		http.Error(rw, msgUnauthorized, http.StatusUnauthorized)
		return
	}

	sessions, err := s.repo.SelectSessionsByPlayerId(c.Id)
	if err != nil {
		http.Error(rw, msgDatabaseError, http.StatusInternalServerError)
		return
	}

	if len(sessions) == sessionsThreshold {
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
		if err = s.repo.DeleteSession(sessions[ind].Id); err != nil {
			http.Error(rw, msgDatabaseError, http.StatusInternalServerError)
			return
		}
	}

	s.genSession(rw, c.Id)
}

// If the verification email fails to send, the token insertion is rolled back,
// letting the player try again.
func (s Service) resetPassword(rw http.ResponseWriter, r *http.Request) {
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

	p, err := s.repo.SelectIdentityByEmail(email)
	if err != nil {
		http.Error(rw, msgUnauthorized, http.StatusUnauthorized)
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
		http.Error(rw, msgCannotHash, http.StatusInternalServerError)
		return
	}

	token := randgen.GenId(randgen.SecureIdLen)
	if err = s.repo.InsertPasswordResetToken(token, p.Id, pwdHash); err != nil {
		log.Print(err)
		http.Error(rw, msgTokenConflict, http.StatusConflict)
		return
	}

	url := os.Getenv("PASSWORD_RESET_ENDPOINT") + token
	var buff bytes.Buffer
	if err = s.emails[1].Execute(&buff, tmplData{Name: p.Name, Url: url}); err != nil {
		log.Print(err)
		http.Error(rw, msgCannotSendEmail, http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(deliveryServicePayload{
		From:     emailSender{Email: os.Getenv("EMAIL_FROM")},
		To:       [1]emailReciever{{email}},
		Subject:  "Password Reset",
		Category: "Transactional",
		Html:     buff.String(),
	})
	if err != nil {
		log.Print(err)
		http.Error(rw, msgCannotSendEmail, http.StatusInternalServerError)
		return
	}

	if err = s.sendEmail(body); err != nil {
		log.Print(err)
		http.Error(rw, msgCannotSendEmail, http.StatusInternalServerError)
		// Remove inserted token.
		if err = s.repo.DeletePasswordResetToken(token); err != nil {
			log.Print(err)
		}
	}
}

// verifySignup completes the registration process for players who click the
// verification email link.
//
// The verification process includes the following steps:
//  1. Fetch signup credentials from database using provided token.
//  2. Insert new player record using provided credentials.
//  3. Delete used token.
//  4. Generate session for the player.
func (s Service) verifySignup(rw http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	data, err := s.repo.SelectSignupDataByToken(token)
	if err != nil {
		log.Print(err)
		http.Redirect(rw, r, "/error", http.StatusFound)
		return
	}

	id := randgen.GenId(randgen.IdLen)
	if err = s.repo.InsertPlayer(id, data); err != nil {
		log.Print(err)
		http.Redirect(rw, r, "/error", http.StatusFound)
		return
	}

	if err = s.repo.DeleteSignupToken(token); err != nil {
		log.Print(err)
	}

	s.genSession(rw, id)

	// Redirect to home page after successfull signup.
	http.Redirect(rw, r, "/", http.StatusFound)
}

// verifyResetPassword completes the password reset process by updating the player
// password and deleting the used password_reset_token.
func (s Service) verifyResetPassword(rw http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	c, err := s.repo.SelectCredentialsByResetToken(token)
	if err != nil {
		log.Print(err)
		http.Redirect(rw, r, "/error", http.StatusFound)
		return
	}

	if err = s.repo.UpdatePasswordHash(c.Id, c.PasswordHash); err != nil {
		log.Print(err)
		http.Redirect(rw, r, "/error", http.StatusFound)
		return
	}

	if err = s.repo.DeletePasswordResetToken(token); err != nil {
		log.Print(err)
	}

	// Redirect to signin page after successfull password reset.
	http.Redirect(rw, r, "/signin", http.StatusFound)
}

// genSession inserts a new record in the session table and adds the HTTP-only
// secure cookie to the response.
func (s Service) genSession(rw http.ResponseWriter, playerId string) {
	// Use generated unique string as session value.
	sessionId := randgen.GenId(randgen.SecureIdLen)

	if err := s.repo.InsertSession(sessionId, playerId); err != nil {
		log.Print(err)
		http.Error(rw, msgDatabaseError, http.StatusInternalServerError)
		return
	}

	http.SetCookie(rw, &http.Cookie{
		Name:     "Auth",
		Value:    sessionId,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // Session will last for 30 days.
		HttpOnly: true,
		// Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// sendEmail sends the email using the Email Delivery Platform.
func (s Service) sendEmail(body []byte) error {
	req, err := http.NewRequest("POST", os.Getenv("EMAIL_SERVICE_URL"), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", os.Getenv("EMAIL_SERVICE_TOKEN"))
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil || (res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent) {
		return errors.New("mailtrap error " + err.Error())
	}
	return res.Body.Close()
}
