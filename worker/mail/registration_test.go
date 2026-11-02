package mail

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/weburl"
)

// withoutCheck adapts a sender of the registration emails, which have no check of the email address.
func withoutCheck(f func(m *Mailer, ctx context.Context, payload []byte) error) sender {
	return func(m *Mailer, ctx context.Context, payload []byte, _ IsEmailAddressValidFunc) error {
		return f(m, ctx, payload)
	}
}

func TestRegistrationNotifications(t *testing.T) {
	tests := []struct {
		name string
		send func(m *Mailer, ctx context.Context, payload []byte) error
		key  string
	}{
		{"RegistrationWelcome", (*Mailer).RegistrationWelcome, "registration.welcome"},
		{"RegistrationConfirmed", (*Mailer).RegistrationConfirmed, "registration.confirmed"},
		{"RegistrationCancelled", (*Mailer).RegistrationCancelled, "registration.cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkNotification(t, withoutCheck(tt.send), tt.key, false)
		})
	}
}

func TestRegistrationConfirmation(t *testing.T) {
	t.Run("sent", func(t *testing.T) {
		mailer, delivered := stub(t)
		if err := mailer.RegistrationConfirmation(t.Context(), encode(t, map[string]string{
			"email": address, "user_id": "uid", "username": "john", "locale": "en", "registration_code": "code",
		})); err != nil {
			t.Fatalf("error = %v", err)
		}

		if len(*delivered) != 1 {
			t.Fatalf("delivered %d messages, want 1", len(*delivered))
		}
		m := (*delivered)[0]
		checkCommon(t, m, "registration.confirm", TemplateConfirmationAndCancel)

		confirmURL := "https://web/en/" + weburl.AccountRegistrationConfirm + "?id=uid&key=code"
		cancelURL := "https://web/en/" + weburl.AccountRegistrationCancel + "?id=uid&key=code"
		if m.HTMLVars["ConfirmURL"] != confirmURL || m.HTMLVars["CancelURL"] != cancelURL {
			t.Errorf("URLs = %v, %v", m.HTMLVars["ConfirmURL"], m.HTMLVars["CancelURL"])
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		mailer, delivered := stub(t)
		if err := mailer.RegistrationConfirmation(t.Context(), []byte("{")); !broker.IsPermanent(err) {
			t.Errorf("error = %v, want a permanent error", err)
		}

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})
}
