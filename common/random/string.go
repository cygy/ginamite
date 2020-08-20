package random

import "math/rand"

var randomChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// String returns a string composed by X random characters.
func String(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = randomChars[rand.Intn(len(randomChars))]
	}

	return string(b)
}
