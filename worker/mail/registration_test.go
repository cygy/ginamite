package mail

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/weburl"
)

// withoutContext adapts a sender of the registration emails, which have no context and no check.
func withoutContext(f func(payload []byte)) sender {
	return func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) { f(payload) }
}

func TestRegistrationNotifications(t *testing.T) {
	tests := []struct {
		name string
		send func(payload []byte)
		key  string
	}{
		{"RegistrationWelcome", RegistrationWelcome, "registration.welcome"},
		{"RegistrationConfirmed", RegistrationConfirmed, "registration.confirmed"},
		{"RegistrationCancelled", RegistrationCancelled, "registration.cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkNotification(t, withoutContext(tt.send), tt.key, false)
		})
	}
}

func TestRegistrationConfirmation(t *testing.T) {
	t.Run("sent", func(t *testing.T) {
		delivered := stub(t)
		RegistrationConfirmation(encode(t, map[string]string{
			"email": address, "user_id": "uid", "username": "john", "locale": "en", "registration_code": "code",
		}))

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
		delivered := stub(t)
		RegistrationConfirmation([]byte("{"))

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})
}
