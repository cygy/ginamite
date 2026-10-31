package authtoken

import (
	"testing"
	"time"
)

func TestRevoked_AddIsRevokedRemove(t *testing.T) {
	r := NewRevoked()

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
			r.Add(tt.tokens)
			if tt.remove != "" {
				r.Remove(tt.remove)
			}
			if got := r.IsRevoked(tt.token); got != tt.want {
				t.Errorf("IsRevoked(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

func TestRevoked_StartDeletingExpired(t *testing.T) {
	r := NewRevoked()
	r.Add(map[string]time.Time{
		"task-expired": time.Now().Add(-time.Minute),
		"task-valid":   time.Now().Add(time.Hour),
	})
	// The interval is in seconds.
	r.StartDeletingExpired(1)

	deadline := time.Now().Add(5 * time.Second)
	for r.IsRevoked("task-expired") {
		if time.Now().After(deadline) {
			t.Fatal("expired token was not removed by the background task")
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !r.IsRevoked("task-valid") {
		t.Error("valid token was removed")
	}
}
