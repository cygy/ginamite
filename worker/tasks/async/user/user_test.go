package user

import (
	"context"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

var tokens = map[string]time.Time{"token": time.Unix(1700000000, 0)}

// checkTokens checks that the revoked tokens, if any, are invalidated.
func checkTokens(t *testing.T, q *workertest.Queue, want bool) {
	t.Helper()

	m, ok := q.Next()
	if !want {
		if ok {
			t.Errorf("unexpected message %+v", m)
		}

		return
	}

	if !ok || m.Topic != queue.TopicCache || m.Type != queue.MessageCacheInvalidAuthToken {
		t.Errorf("message = %+v, %v", m, ok)
	}
}

func TestDeleteByID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		revoked map[string]time.Time
		want    string // The ID of the user passed to the function, if called.
	}{
		{"revoked tokens", `{"user_id":"u1"}`, tokens, "u1"},
		{"no revoked token", `{"user_id":"u2"}`, nil, "u2"},
		{"invalid payload", `{`, tokens, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := workertest.CaptureQueue(t)

			got := ""
			DeleteByID([]byte(tt.payload), func(_ context.Context, userID string) map[string]time.Time {
				got = userID
				return tt.revoked
			})

			if got != tt.want {
				t.Errorf("user = %q, want %q", got, tt.want)
			}
			checkTokens(t, q, got != "" && tt.revoked != nil)
		})
	}

	t.Run("undefined function", func(t *testing.T) {
		DeleteByID([]byte(`{"user_id":"u"}`), nil)
	})
}

func TestDisableByID(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		revoked map[string]time.Time
		want    string
	}{
		{"revoked tokens", `{"user_id":"u1"}`, tokens, "u1"},
		{"no revoked token", `{"user_id":"u2"}`, nil, "u2"},
		{"invalid payload", `{`, tokens, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := workertest.CaptureQueue(t)

			got := ""
			DisableByID([]byte(tt.payload), func(_ context.Context, userID string) map[string]time.Time {
				got = userID
				return tt.revoked
			})

			if got != tt.want {
				t.Errorf("user = %q, want %q", got, tt.want)
			}
			checkTokens(t, q, got != "" && tt.revoked != nil)
		})
	}

	t.Run("undefined function", func(t *testing.T) {
		DisableByID([]byte(`{"user_id":"u"}`), nil)
	})
}

func TestUserTasks(t *testing.T) {
	tests := []struct {
		name string
		run  func(payload []byte, f func(ctx context.Context, userID string))
	}{
		{"EnableByID", EnableByID},
		{"UpdateSocialNetworksByID", UpdateSocialNetworksByID},
		{"RegistrationDone", RegistrationDone},
		{"RegistrationValidated", RegistrationValidated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, calls := "", 0
			record := func(_ context.Context, userID string) { got, calls = userID, calls+1 }

			tt.run([]byte(`{"user_id":"u1"}`), record)
			if got != "u1" || calls != 1 {
				t.Errorf("user = %q, calls = %d", got, calls)
			}

			tt.run([]byte(`{`), record)
			if calls != 1 {
				t.Error("the function must not be called for an invalid payload")
			}

			tt.run([]byte(`{"user_id":"u1"}`), nil)
		})
	}
}
