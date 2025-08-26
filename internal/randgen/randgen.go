package randgen

import (
	"math/rand/v2"
	"strings"
)

const idLength = 8
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

/*
GenBase62 generates a pseudo-random unique string of [idLength] symbols.
*/
func GenBase62() string {
	var b strings.Builder

	for range idLength {
		b.WriteByte(alphabet[rand.IntN(62)])
	}

	return b.String()
}
