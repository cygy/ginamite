package notification

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

func TestNewLogin(t *testing.T) {
	t.Run("email", func(t *testing.T) {
		q := workertest.CaptureQueue(t)

		var gotUser, gotType string
		NewLogin(context.Background(), q.Sender, []byte(`{"user_id":"u1","email":"john@example.com","ip":"1.2.3.4"}`),
			func(_ context.Context, userID, notificationType string) ([]notifications.Target, error) {
				gotUser, gotType = userID, notificationType
				return []notifications.Target{{Target: notifications.TargetEmail}, {Target: notifications.TargetChrome}}, nil
			})

		if gotUser != "u1" || gotType != notifications.TypeNewLogin {
			t.Errorf("targets requested for %q, %q", gotUser, gotType)
		}

		m, ok := q.Next()
		if !ok || m.Topic != queue.TopicMail || m.Type != queue.MessageMailNewLogin || m.Payload["email"] != "john@example.com" || m.Payload["ip"] != "1.2.3.4" {
			t.Errorf("message = %+v, %v", m, ok)
		}
		q.Empty()
	})

	t.Run("invalid payload", func(t *testing.T) {
		NewLogin(context.Background(), nil, []byte("{"), func(context.Context, string, string) ([]notifications.Target, error) {
			t.Error("the targets must not be requested")
			return nil, nil
		})
	})
}
