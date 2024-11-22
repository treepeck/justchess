package jwt_auth

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Subject provides the id and role of the user, encoded into JWT.
type Subject struct {
	Id uuid.UUID `json:"id"`
	R  Role      `json:"r"`
}

// GeneratePair generates a pair of JWT`s: access token and refresh token.
func GeneratePair(s Subject) (at, rt string, err error) {
	// generate and sign access token
	at, err = GenerateToken(s, "ACCESS_TOKEN_SECRET", time.Minute*30)
	if err != nil {
		slog.Warn("can not generate access token", "err", err)
		return
	}
	// generate and sign refresh token
	rt, err = GenerateToken(s, "REFRESH_TOKEN_SECRET", (time.Hour*24)*15)
	if err != nil {
		slog.Warn("can not generate refresh token", "err", err)
		return
	}
	return
}

// GenerateToken generates a token with the specified parameters.
func GenerateToken(s Subject, secret string,
	d time.Duration) (t string, err error) {
	p, err := json.Marshal(s)
	if err != nil {
		return
	}
	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   string(p),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
	})
	t, err = unsigned.SignedString([]byte(os.Getenv(secret)))
	return
}

// Decodes the provided token using the secret.
// If the token is not valid, returns an error.
func DecodeToken(et, secret string) (dt *jwt.Token, err error) {
	dt, err = jwt.ParseWithClaims(et, &jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv(secret)), nil
		},
	)
	return
}

type Role int

const (
	// Guests do not have access to rating system,
	// their data is not stored in a database.
	Guest Role = iota
	// Players can have rating, change profile data.
	Player
)
