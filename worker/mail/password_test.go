package mail

import (
	"testing"

	"github.com/cygy/ginamite/common/weburl"
)

func TestPasswordNotifications(t *testing.T) {
	tests := []struct {
		name string
		send sender
		key  string
	}{
		{"ForgotPasswordConfirmed", ForgotPasswordConfirmed, "forgot_password.confirmed"},
		{"ForgotPasswordCancelled", ForgotPasswordCancelled, "forgot_password.cancelled"},
		{"UpdatePasswordConfirmed", UpdatePasswordConfirmed, "update_password.confirmed"},
		{"UpdatePasswordCancelled", UpdatePasswordCancelled, "update_password.cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkNotification(t, tt.send, tt.key, true)
		})
	}
}

func TestPasswordConfirmations(t *testing.T) {
	tests := []struct {
		name            string
		send            sender
		key             string
		confirm, cancel string
	}{
		{"ForgotPasswordConfirmation", ForgotPasswordConfirmation, "forgot_password.confirm", weburl.AccountForgotPasswordConfirm, weburl.AccountForgotPasswordCancel},
		{"UpdatePasswordConfirmation", UpdatePasswordConfirmation, "update_password.confirm", weburl.AccountUpdatePasswordConfirm, weburl.AccountUpdatePasswordCancel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkProcess(t, tt.send, tt.key, tt.confirm, tt.cancel)
		})
	}
}
