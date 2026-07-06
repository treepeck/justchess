// Heavily inspired by github.com/gorilla/securecookie package.

package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
)

// secureCookie stores the player's id and role. It serves as a session and is
// entirely client-based (not stored anywhere outside of client's Cookie storage).
type secureCookie struct {
	hashKey []byte
	Id      string `json:"i"`
	maxAge  int64
	IsGuest bool `json:"g"`
}

func (c *secureCookie) encrypt() (string, error) {
	// Serialize the credentials.
	credentials, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	// Generate the MAC of data.
	timestamp := time.Now().UTC().Unix()
	mac := c.sign(credentials, timestamp)
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

func (c *secureCookie) decrypt(encrypted []byte) error {
	if len(encrypted) > maxCookieLen {
		return errTooLarge
	}
	decoded, err := base64.RawURLEncoding.DecodeString(string(encrypted))
	if err != nil {
		return errInvalidEncoding
	}
	// WARN: It is important to use SplitN with parameter 3 to correctly
	// parse the expected MAC.
	parts := bytes.SplitN(decoded, []byte("|"), 3)
	if len(parts) != 3 {
		return errInvalidEncrypted
	}
	// Handle the possible expiration.
	createdAt, err := strconv.ParseInt(string(parts[1]), 10, 64)
	if err != nil {
		return errInvalidTimestamp
	}
	now := time.Now().UTC().Unix()
	if createdAt < now-c.maxAge {
		return errExpired
	}
	// Validate MAC.
	expectedMac := c.sign(parts[0], createdAt)
	if subtle.ConstantTimeCompare(expectedMac, parts[2]) != 1 {
		return errUnsigned
	}
	// Deserialize the credentials.
	return json.Unmarshal(parts[0], c)
}

func (c *secureCookie) sign(credentials []byte, timestamp int64) []byte {
	// Concatenate credentials with timestamp
	data := fmt.Sprintf("%s|%b", string(credentials), timestamp)
	h := hmac.New(sha256.New, c.hashKey)
	h.Write([]byte(data))
	return h.Sum(nil)
}
