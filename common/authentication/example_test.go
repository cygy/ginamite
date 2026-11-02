package authentication_test

import (
	"fmt"

	"github.com/cygy/ginamite/common/authentication"
)

// A password is stored as its Argon2id hash, then compared with the password given to sign in.
func ExampleComparePassword() {
	hash, err := authentication.EncryptPassword("correct horse battery staple")
	if err != nil {
		panic(err)
	}

	fmt.Println(authentication.ComparePassword("correct horse battery staple", hash))
	fmt.Println(authentication.ComparePassword("wrong", hash))
	fmt.Println(authentication.NeedsRehash(hash))
	// Output:
	// true
	// false
	// false
}
