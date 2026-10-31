package authentication

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/cygy/ginamite/common/random"

	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
	"golang.org/x/crypto/argon2"
)

// Parameters of Argon2id, the algorithm of the new hashes: the memory is in KiB. They are the minimum recommended by
// OWASP; raise them as the servers allow, NeedsRehash reports the hashes which use older ones.
const (
	argon2Memory      = 19 * 1024
	argon2Iterations  = 2
	argon2Parallelism = 1
	argon2SaltLength  = 16
	argon2KeyLength   = 32

	argon2Prefix = "$argon2id$"
	legacyPrefix = "$6$"
)

// EncryptPassword returns the Argon2id hash of a password, using a random salt, in the PHC format:
// $argon2id$v=19$m=19456,t=2,p=1$salt$hash.
func EncryptPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("unable to generate the salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)

	return fmt.Sprintf("%sv=%d$m=%d,t=%d,p=%d$%s$%s", argon2Prefix, argon2.Version, argon2Memory, argon2Iterations, argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

// ComparePassword reports whether the password matches the encrypted password, an Argon2id hash or a legacy SHA-512
// crypt hash, see NeedsRehash. The comparison takes the same time whatever the first different byte.
func ComparePassword(password, encryptedPassword string) bool {
	switch {
	case strings.HasPrefix(encryptedPassword, argon2Prefix):
		return compareArgon2(password, encryptedPassword)
	case strings.HasPrefix(encryptedPassword, legacyPrefix):
		return compareLegacy(password, encryptedPassword)
	}

	return false
}

// NeedsRehash reports whether the encrypted password should be encrypted again with EncryptPassword, once the user
// signs in and their password is known: it is a legacy hash, or an Argon2id hash with other parameters than the
// current ones.
func NeedsRehash(encryptedPassword string) bool {
	if !strings.HasPrefix(encryptedPassword, argon2Prefix) {
		return strings.HasPrefix(encryptedPassword, legacyPrefix)
	}

	memory, iterations, parallelism, _, _, ok := parseArgon2(encryptedPassword)

	return !ok || memory != argon2Memory || iterations != argon2Iterations || parallelism != argon2Parallelism
}

// DummyCompare spends the time of a password comparison, so that signing in with an unknown account is not faster than
// with a known one.
func DummyCompare(password string) {
	ComparePassword(password, dummyHash())
}

// dummyHash returns a valid hash of an unknown password, computed once.
var dummyHash = sync.OnceValue(func() string {
	hash, err := EncryptPassword(random.String(32))
	if err != nil {
		panic(err)
	}

	return hash
})

// parseArgon2 splits a hash in the PHC format.
func parseArgon2(hash string) (memory, iterations uint32, parallelism uint8, salt, key []byte, ok bool) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return 0, 0, 0, nil, nil, false
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return 0, 0, 0, nil, nil, false
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return 0, 0, 0, nil, nil, false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return 0, 0, 0, nil, nil, false
	}
	key, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return 0, 0, 0, nil, nil, false
	}

	return memory, iterations, parallelism, salt, key, true
}

func compareArgon2(password, hash string) bool {
	memory, iterations, parallelism, salt, key, ok := parseArgon2(hash)
	// A hash with huge parameters must not make the server compute for long or allocate gigabytes.
	if !ok || memory > 4*argon2Memory || iterations > 10 || parallelism == 0 {
		return false
	}

	other := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(key)))

	return subtle.ConstantTimeCompare(key, other) == 1
}

func compareLegacy(password, encryptedPassword string) bool {
	lastIndex := strings.LastIndex(encryptedPassword, "$")
	if lastIndex < 0 {
		return false
	}

	salt := encryptedPassword[:lastIndex]
	hash, err := sha512_crypt.New().Generate([]byte(password), []byte(salt+"$"))

	return err == nil && subtle.ConstantTimeCompare([]byte(hash), []byte(encryptedPassword)) == 1
}
