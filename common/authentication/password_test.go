package authentication

import (
	"strings"
	"testing"

	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

// legacyHash returns the SHA-512 crypt hash the first versions of the package created.
func legacyHash(t *testing.T, password string) string {
	t.Helper()

	hash, err := sha512_crypt.New().Generate([]byte(password), []byte("$6$rounds=5000$abcdefghijklmnop$"))
	if err != nil {
		t.Fatal(err)
	}

	return hash
}

func TestEncryptPassword(t *testing.T) {
	first, err := EncryptPassword("Secret123456")
	if err != nil {
		t.Fatalf("EncryptPassword() error = %v", err)
	}

	second, err := EncryptPassword("Secret123456")
	if err != nil {
		t.Fatalf("EncryptPassword() error = %v", err)
	}

	if !strings.HasPrefix(first, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Errorf("hash = %q, want an Argon2id hash", first)
	}
	if strings.Contains(first, "Secret123456") || first == second {
		t.Error("the hash must not contain the password and must use a random salt")
	}
}

func TestComparePassword(t *testing.T) {
	hash, err := EncryptPassword("Secret123456")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"match", "Secret123456", hash, true},
		{"legacy match", "Secret123456", legacyHash(t, "Secret123456"), true},
		{"legacy wrong password", "secret123456", legacyHash(t, "Secret123456"), false},
		{"wrong password", "secret123456", hash, false},
		{"empty password", "", hash, false},
		{"empty hash", "Secret123456", "", false},
		{"unknown format", "Secret123456", "nodollar", false},
		{"truncated hash", "Secret123456", hash[:len(hash)-2], false},
		{"truncated legacy hash", "Secret123456", legacyHash(t, "Secret123456")[:20], false},
		{"bad version", "Secret123456", strings.Replace(hash, "v=19", "v=16", 1), false},
		{"bad parameters", "Secret123456", strings.Replace(hash, "m=19456,t=2,p=1", "m=x", 1), false},
		{"bad salt", "Secret123456", strings.Replace(hash, "p=1$", "p=1$!!", 1), false},
		{"too much memory", "Secret123456", strings.Replace(hash, "m=19456", "m=4000000", 1), false},
		{"too many iterations", "Secret123456", strings.Replace(hash, "t=2", "t=999", 1), false},
		{"no parallelism", "Secret123456", strings.Replace(hash, "p=1", "p=0", 1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComparePassword(tt.password, tt.hash); got != tt.want {
				t.Errorf("ComparePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNeedsRehash(t *testing.T) {
	current, err := EncryptPassword("Secret123456")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		hash string
		want bool
	}{
		{"current", current, false},
		{"legacy", legacyHash(t, "Secret123456"), true},
		{"older parameters", strings.Replace(current, "t=2", "t=1", 1), true},
		{"invalid argon2", "$argon2id$garbage", true},
		{"unknown", "plain", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NeedsRehash(tt.hash); got != tt.want {
				t.Errorf("NeedsRehash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummyCompare(t *testing.T) {
	DummyCompare("anything")

	if !strings.HasPrefix(dummyHash(), "$argon2id$") || ComparePassword("anything", dummyHash()) {
		t.Error("the dummy hash must be a valid hash of an unknown password")
	}
}
