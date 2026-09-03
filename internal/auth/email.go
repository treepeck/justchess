package auth

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"os"
)

// tmplData is used to fill up the email templates.
type tmplData struct {
	Name string
	Url  string
}

// sender is a nested object needed to form a valid [email].
type sender struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// reciever is a nested object needed to form a valid [email].
type reciever struct {
	Email string `json:"email"`
}

// emailPayload is needed to send email through a Mailtrap service.
type emailPayload struct {
	From     sender      `json:"from"`
	To       [1]reciever `json:"to"`
	Subject  string      `json:"subject"`
	Html     string      `json:"html"`
	Category string      `json:"category"`
}

func (s *Service) ParseEmails(folder string) error {
	signup, err := template.ParseFiles(folder + "email-signup.tmpl")
	if err != nil {
		return err
	}
	s.emails[0] = signup

	reset, err := template.ParseFiles(folder + "email-reset.tmpl")
	if err != nil {
		return err
	}
	s.emails[1] = reset
	return nil
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
