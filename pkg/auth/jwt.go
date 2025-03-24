package auth

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CtxKey string

const (
	Subj CtxKey = "subj"
)

type Subject struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// generateToken encodes a token using the provided secret string.
// Uses Subject type as a token subject.
// If token cannot be signed, returns error.
func generateToken(id uuid.UUID, name, secret string, d time.Duration) (string, error) {
	s := Subject{
		Id:   id,
		Name: name,
	}
	payload, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   string(payload),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
	})
	return unsigned.SignedString([]byte(secret))
}

// generatePair generates a pair of JWTs: access token and refresh token.
func generatePair(id uuid.UUID, name string) (at, rt string, err error) {
	at, err = generateToken(id, name, os.Getenv("ACCESS_TOKEN_SECRET"), time.Minute*30)
	if err != nil {
		log.Printf("cannot generate access token: %v\n", err)
		return
	}
	rt, err = generateToken(id, name, os.Getenv("REFRESH_TOKEN_SECRET"), (time.Hour*24)*30)
	if err != nil {
		log.Printf("cannot generate refresh token: %v\n", err)
		return
	}
	return
}

// DecodeToken decodes the provided token using the provided secret.
// Retunts decode subject, if the token is valid.
// If the token is not valid, returns an error.
// secret param specifies the key of the environment variable.
func DecodeToken(encoded, secret string) (s Subject, err error) {
	t, err := jwt.ParseWithClaims(encoded, &jwt.RegisteredClaims{},
		func(t *jwt.Token) (any, error) {
			return []byte(os.Getenv(secret)), nil
		},
	)
	if err != nil {
		return
	}

	subj, err := t.Claims.GetSubject()
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(subj), &s)
	return
}
