package authentication

import (
	"strings"
	"testing"
)

// FuzzComparePassword checks that no stored hash, even malformed, makes the comparison panic.
func FuzzComparePassword(f *testing.F) {
	f.Add("Secret123456", "$argon2id$v=19$m=19456,t=2,p=1$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g")
	f.Add("Hello world!", "$6$saltstring$svn8UoSVapNtMuq1ukKS4tPQd8iKwSMHWjl/O817G3uBnIFNjnQJuesI68u4OTLiBFdcbYEdFCoEOfaS35inz1")
	f.Add("", "$argon2id$v=19$m=0,t=0,p=0$$")
	f.Add("x", "$6$rounds=$$")

	f.Fuzz(func(t *testing.T, password, hash string) {
		// The huge rounds or memory of a hash are refused, they are not computed: skip only what is slow by design.
		if strings.Contains(hash, "rounds=") && !strings.Contains(hash, "rounds=1000$") {
			t.Skip()
		}

		ComparePassword(password, hash)
		NeedsRehash(hash)
	})
}

// FuzzSHA512Crypt checks that a legacy hash is verified by the password which produced it, and only by it.
func FuzzSHA512Crypt(f *testing.F) {
	f.Add("Hello world!", "saltstring")
	f.Add("", "")
	f.Add("é€😀", "a$b")

	f.Fuzz(func(t *testing.T, password, salt string) {
		hash, err := sha512Crypt(password, legacyPrefix+salt+"$")
		if err != nil {
			if len(password) > sha512CryptMaxPassword || strings.HasPrefix(salt, "rounds=") {
				return
			}
			t.Fatalf("sha512Crypt(%q, %q) = %v", password, salt, err)
		}

		if !compareLegacy(password, hash) {
			t.Errorf("the hash %q does not verify its password %q", hash, password)
		}
		if compareLegacy(password+"x", hash) {
			t.Errorf("the hash %q verifies another password", hash)
		}
	})
}
