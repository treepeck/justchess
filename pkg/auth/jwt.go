package auth

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// generatePair generates a pair of JWT`s: access token and refresh token.
func generatePair(id uuid.UUID) (at, rt string, err error) {
	at, err = generateToken(id, "ATS", time.Minute*30)
	if err != nil {
		slog.Error("Cannot generate access token.", "err", err)
		return
	}
	rt, err = generateToken(id, "RTS", (time.Hour*24)*30)
	if err != nil {
		slog.Error("Cannot generate refresh token.", "err", err)
		return
	}
	return
}

// generateToken encodes a token using the provided secret string.
// If token cannot be signed, returns error.
func generateToken(id uuid.UUID, secret string,
	d time.Duration) (t string, err error) {
	s := os.Getenv(secret)
	if s == "" {
		return "", errors.New("cannot read evironment variable")
	}

	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   id.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
	})
	t, err = unsigned.SignedString([]byte(s))
	return
}

// DecodeToken decodes the provided token using the provided secret.
// If the token is not valid, returns an error.
func DecodeToken(et, secret string) (dt *jwt.Token, err error) {
	s := os.Getenv(secret)
	if s == "" {
		return nil, errors.New("cannot read environment variable")
	}

	dt, err = jwt.ParseWithClaims(et, &jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(s), nil
		},
	)
	return
}
