package authentication

import (
	"fmt"
	"strings"

	"github.com/cygy/ginamite/common/random"

	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

// EncryptPassword returns the SHA-512 crypt hash of a password, using a random salt.
func EncryptPassword(password string) (string, error) {
	randomSalt := random.String(PasswordEncryptionSaltLength)
	salt := fmt.Sprintf("%s%s$", PasswordEncryptionSalt, randomSalt)
	crypt := sha512_crypt.New()

	return crypt.Generate([]byte(password), []byte(salt))
}

// ComparePassword reports whether the password matches the encrypted password.
func ComparePassword(password, encryptedPassword string) bool {
	lastIndex := strings.LastIndex(encryptedPassword, "$")
	if lastIndex < 0 {
		return false
	}

	salt := encryptedPassword[:lastIndex]
	crypt := sha512_crypt.New()
	hash, err := crypt.Generate([]byte(password), []byte(salt+"$"))

	return err == nil && hash == encryptedPassword
}
