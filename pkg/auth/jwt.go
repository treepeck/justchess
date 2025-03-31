package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Role int
type CtxKey string

const (
	RoleGuest Role = iota
	RoleUser
	Cms CtxKey = "claims" // Claims context key.
)

type Claims struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Role Role      `json:"role"`
	jwt.RegisteredClaims
}

// generateToken encodes a token using the provided secret string.
// If token cannot be signed, returns error.
func generateToken(id uuid.UUID, name string, r Role, secret string, d time.Duration) (string, error) {
	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{id, name, r,
		jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(d))},
	})
	return unsigned.SignedString([]byte(os.Getenv(secret)))
}

// generatePair generates a pair of JWTs: access token and refresh token.
func generatePair(id uuid.UUID, name string, r Role) (at, rt string, err error) {
	at, err = generateToken(id, name, r, "ACCESS_TOKEN_SECRET", time.Minute*30)
	if err != nil {
		log.Printf("cannot generate access token: %v\n", err)
		return
	}

	rt, err = generateToken(id, name, r, "REFRESH_TOKEN_SECRET", time.Hour*24*30)
	if err != nil {
		log.Printf("cannot generate refresh token: %v\n", err)
	}
	return
}

// DecodeToken decodes the provided token using the provided secret.
// Retunts decode subject, if the token is valid.
// If the token is not valid, returns an error.
// secret param specifies the key of the environment variable.
func DecodeToken(encoded, secret string) (Claims, error) {
	t, err := jwt.ParseWithClaims(encoded, &Claims{},
		func(t *jwt.Token) (any, error) {
			return []byte(os.Getenv(secret)), nil
		},
	)
	if err != nil {
		return Claims{}, err
	}

	c, ok := t.Claims.(*Claims)
	if !ok {
		err = errors.New("cannot parse claims")
	}
	return *c, err
}
