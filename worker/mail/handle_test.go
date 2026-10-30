package mail

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/queue"
)

func TestHandle(t *testing.T) {
	ctx := context.Background()

	t.Run("built-in messages", func(t *testing.T) {
		for messageType := range handlers {
			delivered := stub(t)

			// An invalid payload is enough to know the message is handled.
			if !Handle(ctx, messageType, []byte("{"), valid) {
				t.Errorf("%s is not handled", messageType)
			}
			if len(*delivered) != 0 {
				t.Errorf("%s: delivered %d messages, want none", messageType, len(*delivered))
			}
		}
	})

	t.Run("delivers the email", func(t *testing.T) {
		delivered := stub(t)

		if !Handle(ctx, queue.MessageMailRegistrationWelcome, notificationPayload(t), nil) {
			t.Fatal("the message is not handled")
		}
		if len(*delivered) != 1 {
			t.Errorf("delivered %d messages, want 1", len(*delivered))
		}
	})

	t.Run("unknown message", func(t *testing.T) {
		if Handle(ctx, "unknown", nil, valid) {
			t.Error("an unknown message must not be handled")
		}
	})
}
