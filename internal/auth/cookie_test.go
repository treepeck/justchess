package auth

import (
	"testing"
	"justchess/internal/randgen"
	"crypto/rand"
)

func BenchmarkEncrypt(b *testing.B) {
	key := make([]byte, 64)
	rand.Read(key)

	c := &secureCookie{
		Id: randgen.GenId(randgen.IdLen),
		IsGuest: false,
		hashKey: key,
		maxAge: playerSessionMaxAge,
	}
	for b.Loop() {
		if _, err := c.encrypt(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := make([]byte, 64)
	rand.Read(key)

	c := &secureCookie{
		Id: randgen.GenId(randgen.IdLen),
		IsGuest: false,
		hashKey: key,
		maxAge: playerSessionMaxAge,
	}
	raw, err := c.encrypt()
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if err = c.decrypt([]byte(raw)); err != nil {
			b.Fatal(err)
		}
	}
}
