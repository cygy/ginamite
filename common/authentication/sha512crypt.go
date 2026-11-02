package authentication

import (
	"crypto/sha512"
	"errors"
	"strconv"
	"strings"
)

// The parameters of SHA-512 crypt, see https://www.akkadia.org/drepper/SHA-crypt.txt.
const (
	sha512CryptDefaultRounds = 5000
	sha512CryptMinRounds     = 1000
	// sha512CryptMaxRounds is far below the 999999999 of the specification: a stored hash must not make the server
	// compute for long.
	sha512CryptMaxRounds = 1_000_000
	sha512CryptMaxSalt   = 16
	sha512CryptAlphabet  = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// errInvalidSHA512Crypt is returned for a setting which is not a SHA-512 crypt one, or with too many rounds.
var errInvalidSHA512Crypt = errors.New("invalid SHA-512 crypt setting")

// sha512CryptPermutation is the order in which the bytes of the digest are encoded, three at a time.
var sha512CryptPermutation = [21][3]int{
	{0, 21, 42},
	{22, 43, 1},
	{44, 2, 23},
	{3, 24, 45},
	{25, 46, 4},
	{47, 5, 26},
	{6, 27, 48},
	{28, 49, 7},
	{50, 8, 29},
	{9, 30, 51},
	{31, 52, 10},
	{53, 11, 32},
	{12, 33, 54},
	{34, 55, 13},
	{56, 14, 35},
	{15, 36, 57},
	{37, 58, 16},
	{59, 17, 38},
	{18, 39, 60},
	{40, 61, 19},
	{62, 20, 41},
}

// sha512Crypt returns the SHA-512 crypt hash of the password with the setting, a hash or its prefix such as
// "$6$rounds=5000$salt$". Only the legacy hashes are verified with it: the new ones use Argon2id.
func sha512Crypt(password, setting string) (string, error) {
	rest, ok := strings.CutPrefix(setting, legacyPrefix)
	if !ok {
		return "", errInvalidSHA512Crypt
	}

	rounds, customRounds := sha512CryptDefaultRounds, false
	if value, after, found := strings.Cut(rest, "$"); found && strings.HasPrefix(value, "rounds=") {
		n, err := strconv.Atoi(strings.TrimPrefix(value, "rounds="))
		if err != nil || n > sha512CryptMaxRounds {
			return "", errInvalidSHA512Crypt
		}
		rounds, customRounds, rest = max(n, sha512CryptMinRounds), true, after
	}

	salt, _, _ := strings.Cut(rest, "$")
	if len(salt) > sha512CryptMaxSalt {
		salt = salt[:sha512CryptMaxSalt]
	}

	digest := sha512CryptDigest([]byte(password), []byte(salt), rounds)

	var b strings.Builder
	b.WriteString(legacyPrefix)
	if customRounds {
		b.WriteString("rounds=" + strconv.Itoa(rounds) + "$")
	}
	b.WriteString(salt + "$")
	for _, p := range sha512CryptPermutation {
		writeSHA512CryptBase64(&b, uint(digest[p[0]])<<16|uint(digest[p[1]])<<8|uint(digest[p[2]]), 4)
	}
	writeSHA512CryptBase64(&b, uint(digest[63]), 2)

	return b.String(), nil
}

// sha512CryptDigest computes the digest of the password and the salt with the rounds, as the specification says.
func sha512CryptDigest(password, salt []byte, rounds int) []byte {
	alternate := sha512.New()
	alternate.Write(password)
	alternate.Write(salt)
	alternate.Write(password)
	alternateSum := alternate.Sum(nil)

	a := sha512.New()
	a.Write(password)
	a.Write(salt)
	a.Write(repeat(alternateSum, len(password)))
	for n := len(password); n > 0; n >>= 1 {
		if n&1 == 1 {
			a.Write(alternateSum)
		} else {
			a.Write(password)
		}
	}
	sum := a.Sum(nil)

	dp := sha512.New()
	for range password {
		dp.Write(password)
	}
	p := repeat(dp.Sum(nil), len(password))

	ds := sha512.New()
	for range 16 + int(sum[0]) {
		ds.Write(salt)
	}
	s := repeat(ds.Sum(nil), len(salt))

	for i := range rounds {
		c := sha512.New()
		if i&1 == 1 {
			c.Write(p)
		} else {
			c.Write(sum)
		}
		if i%3 != 0 {
			c.Write(s)
		}
		if i%7 != 0 {
			c.Write(p)
		}
		if i&1 == 1 {
			c.Write(sum)
		} else {
			c.Write(p)
		}
		sum = c.Sum(sum[:0])
	}

	return sum
}

// repeat returns length bytes made of b repeated.
func repeat(b []byte, length int) []byte {
	out := make([]byte, 0, length)
	for len(out) < length {
		out = append(out, b[:min(len(b), length-len(out))]...)
	}

	return out
}

// writeSHA512CryptBase64 writes the n lowest groups of 6 bits of w, the lowest first.
func writeSHA512CryptBase64(b *strings.Builder, w uint, n int) {
	for range n {
		b.WriteByte(sha512CryptAlphabet[w&0x3f])
		w >>= 6
	}
}
