package authtoken

import (
	"testing"
	"time"
)

func TestStartDeletingExpired_RemovesExpiredTokens(t *testing.T) {
	Add(map[string]time.Time{
		"task-expired": time.Now().Add(-time.Minute),
		"task-valid":   time.Now().Add(time.Hour),
	})
	t.Cleanup(func() { Remove("task-valid") })

	// The interval is in seconds.
	StartDeletingExpired(1)

	deadline := time.Now().Add(5 * time.Second)
	for IsRevoked("task-expired") {
		if time.Now().After(deadline) {
			t.Fatal("expired token was not removed by the background task")
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !IsRevoked("task-valid") {
		t.Error("valid token was removed")
	}
}
