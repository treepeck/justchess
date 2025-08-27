package randgen

import (
	"crypto/rand"
	"encoding/base64"
)

/*
Final encoded player id will have length 12, session id - 32.
See https://en.wikipedia.org/wiki/Base64
*/
const (
	PlayerIdLen  int = 9
	SessionIdLen int = 24
)

/*
GenId generates a secure random array of n bytes and applies a base64 encoding
for storing in cookies and url.
*/
func GenId(n int) string {
	buff := make([]byte, n)

	// Read never returns an error, so omit the check.
	rand.Read(buff)

	return base64.RawURLEncoding.EncodeToString(buff)
}
