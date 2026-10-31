package mail

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/queue"
)

func TestHandle(t *testing.T) {
	ctx := context.Background()

	t.Run("built-in messages", func(t *testing.T) {
		for messageType := range new(Mailer).handlers() {
			mailer, delivered := stub(t)

			// An invalid payload is enough to know the message is handled.
			handled, err := mailer.Handle(ctx, messageType, []byte("{"), valid)
			if !handled || !broker.IsPermanent(err) {
				t.Errorf("%s: handled = %v, error = %v, want a handled message and a permanent error", messageType, handled, err)
			}
			if len(*delivered) != 0 {
				t.Errorf("%s: delivered %d messages, want none", messageType, len(*delivered))
			}
		}
	})

	t.Run("delivers the email", func(t *testing.T) {
		mailer, delivered := stub(t)

		if handled, err := mailer.Handle(ctx, queue.MessageMailRegistrationWelcome, notificationPayload(t), nil); !handled || err != nil {
			t.Fatalf("handled = %v, error = %v", handled, err)
		}
		if len(*delivered) != 1 {
			t.Errorf("delivered %d messages, want 1", len(*delivered))
		}
	})

	t.Run("unknown message", func(t *testing.T) {
		mailer, _ := stub(t)
		if handled, err := mailer.Handle(ctx, "unknown", nil, valid); handled || err != nil {
			t.Errorf("handled = %v, error = %v, an unknown message must not be handled", handled, err)
		}
	})
}
