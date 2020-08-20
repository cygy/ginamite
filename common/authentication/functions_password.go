package authentication

import (
	"fmt"
	"strings"

	"github.com/cygy/ginamite/common/random"

	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

// EncryptPassword : encrypt a password.
func EncryptPassword(password string) (string, error) {
	randomSalt := random.String(PasswordEncryptionSaltLength)
	salt := fmt.Sprintf("%s%s$", PasswordEncryptionSalt, randomSalt)
	crypt := sha512_crypt.New()
	hash, err := crypt.Generate([]byte(password), []byte(salt))

	return hash, err
}

// ComparePassword : return true if the password is equal to the encrypted password.
func ComparePassword(password, encryptedPassword string) bool {
	lastIndex := strings.LastIndex(encryptedPassword, "$")
	salt := encryptedPassword[:lastIndex]
	crypt := sha512_crypt.New()
	hash, err := crypt.Generate([]byte(password), []byte(salt+"$"))

	return err == nil && hash == encryptedPassword
}
