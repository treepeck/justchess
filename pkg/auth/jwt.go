package auth

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// generatePair generates a pair of JWTs: access token and refresh token.
// The os.Args (command-line arguments) must store secrets for safe
// token signing.
func generatePair(id uuid.UUID) (at, rt string, err error) {
	at, err = generateToken(id, os.Args[1], time.Minute*30)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	rt, err = generateToken(id, os.Args[2], (time.Hour*24)*30)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	return
}

// generateToken encodes a token using the provided secret string.
// If token cannot be signed, returns error.
func generateToken(id uuid.UUID, secret string,
	d time.Duration) (string, error) {
	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   id.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
	})
	return unsigned.SignedString([]byte(secret))
}

// DecodeToken decodes the provided token using the provided secret.
// If the token is not valid, returns an error.
// secretParam specifies the index of command-line argument.
// 1 - ACCESS_TOKEN_SECRET_STRING;
// 2 - REFRESH_TOKEN_SECRET_STRING.
func DecodeToken(encoded string, secretParam int) (*jwt.Token, error) {
	return jwt.ParseWithClaims(encoded, &jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Args[secretParam]), nil
		},
	)
}
