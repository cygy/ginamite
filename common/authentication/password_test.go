package authentication

import "testing"

func TestEncryptPassword(t *testing.T) {
	first, err := EncryptPassword("Secret123456")
	if err != nil {
		t.Fatalf("EncryptPassword() error = %v", err)
	}

	second, err := EncryptPassword("Secret123456")
	if err != nil {
		t.Fatalf("EncryptPassword() error = %v", err)
	}

	if first == "Secret123456" || first == second {
		t.Error("the hash must not be the password and must use a random salt")
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
		{"wrong password", "secret123456", hash, false},
		{"empty password", "", hash, false},
		{"empty hash", "Secret123456", "", false},
		{"hash without separator", "Secret123456", "nodollar", false},
		{"truncated hash", "Secret123456", hash[:len(hash)-1], false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComparePassword(tt.password, tt.hash); got != tt.want {
				t.Errorf("ComparePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
