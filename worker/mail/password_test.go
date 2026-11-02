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
		{"ForgotPasswordConfirmed", (*Mailer).ForgotPasswordConfirmed, "forgot_password.confirmed"},
		{"ForgotPasswordCancelled", (*Mailer).ForgotPasswordCancelled, "forgot_password.cancelled"},
		{"UpdatePasswordConfirmed", (*Mailer).UpdatePasswordConfirmed, "update_password.confirmed"},
		{"UpdatePasswordCancelled", (*Mailer).UpdatePasswordCancelled, "update_password.cancelled"},
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
		{"ForgotPasswordConfirmation", (*Mailer).ForgotPasswordConfirmation, "forgot_password.confirm", weburl.AccountForgotPasswordConfirm, weburl.AccountForgotPasswordCancel},
		{"UpdatePasswordConfirmation", (*Mailer).UpdatePasswordConfirmation, "update_password.confirm", weburl.AccountUpdatePasswordConfirm, weburl.AccountUpdatePasswordCancel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkProcess(t, tt.send, tt.key, tt.confirm, tt.cancel)
		})
	}
}
