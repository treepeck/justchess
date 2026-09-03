package auth

import (
	"crypto/rand"
	"justchess/internal/randgen"
	"os"
	"testing"
)

func TestParseCookieKey(t *testing.T) {
	raw := os.Getenv("COOKIE_KEY")
	var err error
	if _, err = ParseCookieKey(raw); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkGenSecureCookie(b *testing.B) {
	key := make([]byte, 64)
	rand.Read(key)

	s := Session{Id: randgen.GenId(randgen.IdLen), IsGuest: false}
	for b.Loop() {
		if _, err := genSecureCookie(s, key); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateSecureCookie(b *testing.B) {
	key := make([]byte, 64)
	rand.Read(key)

	s := Session{Id: randgen.GenId(randgen.IdLen), IsGuest: false}
	raw, err := genSecureCookie(s, key)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if s, err = validateSecureCookie([]byte(raw), key); err != nil {
			b.Fatal(err)
		}
	}
}
