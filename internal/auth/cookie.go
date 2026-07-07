// Heavily inspired by github.com/gorilla/securecookie package.

package auth

import (
	"bytes"
	"strings"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

const maxCookieLen = 4000 // In bytes.

var (
	errTooLarge         = errors.New("auth: cookie len limit exceeded")
	errInvalidEncrypted = errors.New("auth: encrypted cookie is invalid")
	errInvalidEncoding  = errors.New("auth: invalid base64 encoding")
	errUnsigned         = errors.New("auth: maybe next time")
	errInvalidTimestamp = errors.New("auth: invalid timestamp")
	errExpired          = errors.New("auth: cookie is expired")
	errKey              = errors.New("auth: cannot parse the cookie hash key")
)

func ParseCookieKey(raw string) ([]byte, error) {
	result := make([]byte, 0, 64)
	for chunk := range strings.SplitSeq(raw, " ") {
		b, err := strconv.ParseInt(chunk, 10, 32)
		if err != nil {
			log.Println(err.(*strconv.NumError).Err)
			return nil, errKey
		}
		result = append(result, byte(b))
	}
	return result, nil
}

// Session stores the player's id and role. It serves as a session and is
// entirely client-based (not stored anywhere outside of client's Cookie storage).
type Session struct {
	Id      string `json:"i"`
	IsGuest bool `json:"g"`
}

func genSecureCookie(s Session, hashKey []byte) (string, error) {
	// Serialize the credentials.
	credentials, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	// Generate the MAC of data.
	timestamp := time.Now().UTC().Unix()
	mac := signSecureCookie(credentials, hashKey, timestamp)
	// Concatenate MAC with data.
	// WARN: It is important to place the MAC in the end of the string, since it can
	// contain '|' symbols. If it will be not the last part of the cookie, the decrypt
	// function will break.
	result := fmt.Sprintf("%s|%d|%s", string(credentials), timestamp, string(mac))
	encoded := base64.RawURLEncoding.EncodeToString([]byte(result))
	if len(encoded) > maxCookieLen {
		return "", errTooLarge
	}
	return encoded, nil
}

func validateSecureCookie(encrypted, hashKey []byte) (Session, error) {
	if len(encrypted) > maxCookieLen {
		return Session{}, errTooLarge
	}
	decoded, err := base64.RawURLEncoding.DecodeString(string(encrypted))
	if err != nil {
		return Session{}, errInvalidEncoding
	}
	// WARN: It is important to use SplitN with parameter 3 to correctly
	// parse the expected MAC.
	parts := bytes.SplitN(decoded, []byte("|"), 3)
	if len(parts) != 3 {
		return Session{}, errInvalidEncrypted
	}
	// Handle the possible expiration.
	createdAt, err := strconv.ParseInt(string(parts[1]), 10, 64)
	if err != nil {
		return Session{}, errInvalidTimestamp
	}
	if createdAt < time.Now().UTC().Unix()-cookieMaxAge {
		return Session{}, errExpired
	}
	// Validate MAC.
	expectedMac := signSecureCookie(parts[0], hashKey, createdAt)
	if subtle.ConstantTimeCompare(expectedMac, parts[2]) != 1 {
		return Session{}, errUnsigned
	}
	// Deserialize the credentials.
	var s Session
	return s, json.Unmarshal(parts[0], &s)
}

func signSecureCookie(session, hashKey []byte, timestamp int64) []byte {
	// Concatenate credentials with timestamp
	data := fmt.Sprintf("%s|%b", string(session), timestamp)
	h := hmac.New(sha256.New, hashKey)
	h.Write([]byte(data))
	return h.Sum(nil)
}
