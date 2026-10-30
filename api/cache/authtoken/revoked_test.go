package authtoken

import (
	"testing"
	"time"
)

func TestRevoked_AddIsRevokedRemove(t *testing.T) {
	tests := []struct {
		name   string
		tokens map[string]time.Time
		remove string
		token  string
		want   bool
	}{
		{"added token is revoked", map[string]time.Time{"revoked-1": time.Now().Add(time.Hour)}, "", "revoked-1", true},
		{"unknown token is not revoked", map[string]time.Time{}, "", "revoked-unknown", false},
		{"removed token is not revoked", map[string]time.Time{"revoked-2": time.Now().Add(time.Hour)}, "revoked-2", "revoked-2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Add(tt.tokens)
			if tt.remove != "" {
				Remove(tt.remove)
			}
			t.Cleanup(func() { revoked.remove(tt.token) })

			if got := IsRevoked(tt.token); got != tt.want {
				t.Errorf("IsRevoked(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}
