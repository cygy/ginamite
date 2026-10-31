package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/kafka"
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
			err := DeleteByID([]byte(tt.payload), q.Sender, func(_ context.Context, userID string) (map[string]time.Time, error) {
				got = userID
				return tt.revoked, nil
			})

			if got != tt.want {
				t.Errorf("user = %q, want %q", got, tt.want)
			}
			if (err != nil) != (tt.want == "") || (err != nil && !kafka.IsPermanent(err)) {
				t.Errorf("error = %v, want a permanent error for an invalid payload only", err)
			}
			checkTokens(t, q, got != "" && tt.revoked != nil)
		})
	}

	t.Run("undefined function", func(t *testing.T) {
		if err := DeleteByID([]byte(`{"user_id":"u"}`), nil, nil); err != nil {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		q := workertest.CaptureQueue(t)

		boom := errors.New("boom")
		err := DeleteByID([]byte(`{"user_id":"u"}`), q.Sender, func(context.Context, string) (map[string]time.Time, error) {
			return tokens, boom
		})

		if !errors.Is(err, boom) || kafka.IsPermanent(err) {
			t.Errorf("error = %v, want the error of the function, which a retry can fix", err)
		}
		checkTokens(t, q, false)
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
			err := DisableByID([]byte(tt.payload), q.Sender, func(_ context.Context, userID string) (map[string]time.Time, error) {
				got = userID
				return tt.revoked, nil
			})

			if got != tt.want {
				t.Errorf("user = %q, want %q", got, tt.want)
			}
			if (err != nil) != (tt.want == "") || (err != nil && !kafka.IsPermanent(err)) {
				t.Errorf("error = %v, want a permanent error for an invalid payload only", err)
			}
			checkTokens(t, q, got != "" && tt.revoked != nil)
		})
	}

	t.Run("undefined function", func(t *testing.T) {
		if err := DisableByID([]byte(`{"user_id":"u"}`), nil, nil); err != nil {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		q := workertest.CaptureQueue(t)

		boom := errors.New("boom")
		err := DisableByID([]byte(`{"user_id":"u"}`), q.Sender, func(context.Context, string) (map[string]time.Time, error) {
			return tokens, boom
		})

		if !errors.Is(err, boom) || kafka.IsPermanent(err) {
			t.Errorf("error = %v, want the error of the function, which a retry can fix", err)
		}
		checkTokens(t, q, false)
	})
}

func TestUserTasks(t *testing.T) {
	tests := []struct {
		name string
		run  func(payload []byte, f func(ctx context.Context, userID string) error) error
	}{
		{"EnableByID", EnableByID},
		{"UpdateSocialNetworksByID", UpdateSocialNetworksByID},
		{"RegistrationDone", RegistrationDone},
		{"RegistrationValidated", RegistrationValidated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, calls := "", 0
			record := func(_ context.Context, userID string) error { got, calls = userID, calls+1; return nil }

			if err := tt.run([]byte(`{"user_id":"u1"}`), record); err != nil || got != "u1" || calls != 1 {
				t.Errorf("user = %q, calls = %d, error = %v", got, calls, err)
			}

			if err := tt.run([]byte(`{`), record); !kafka.IsPermanent(err) || calls != 1 {
				t.Errorf("calls = %d, error = %v: the function must not be called for an invalid payload", calls, err)
			}

			if err := tt.run([]byte(`{"user_id":"u1"}`), nil); err != nil {
				t.Errorf("error = %v for an undefined function", err)
			}

			boom := errors.New("boom")
			if err := tt.run([]byte(`{"user_id":"u1"}`), func(context.Context, string) error { return boom }); !errors.Is(err, boom) || kafka.IsPermanent(err) {
				t.Errorf("error = %v, want the error of the function, which a retry can fix", err)
			}
		})
	}
}
