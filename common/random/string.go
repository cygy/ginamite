// Package random generates random strings.
package random

import (
	"crypto/rand"
	"math/big"
)

var randomChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// String returns a string of n random alphanumeric characters.
// The characters come from a cryptographically secure source, so the result can be used for keys and codes.
// It returns an empty string if n is not positive.
func String(n int) string {
	if n <= 0 {
		return ""
	}

	limit := big.NewInt(int64(len(randomChars)))
	b := make([]rune, n)

	for i := range b {
		index, err := rand.Int(rand.Reader, limit)
		if err != nil {
			// The system random source is unavailable: no safe fallback exists.
			panic("random: unable to read the system random source: " + err.Error())
		}

		b[i] = randomChars[index.Int64()]
	}

	return string(b)
}
