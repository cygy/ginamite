package mail

import (
	"testing"

	"github.com/cygy/ginamite/common/weburl"
)

func TestAccountNotifications(t *testing.T) {
	tests := []struct {
		name string
		send sender
		key  string
	}{
		{"DeleteAccountConfirmed", DeleteAccountConfirmed, "delete_account.confirmed"},
		{"DeleteAccountCancelled", DeleteAccountCancelled, "delete_account.cancelled"},
		{"DisableAccountConfirmed", DisableAccountConfirmed, "disable_account.confirmed"},
		{"DisableAccountCancelled", DisableAccountCancelled, "disable_account.cancelled"},
		{"EnableAccountConfirmed", EnableAccountConfirmed, "enable_account.confirmed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkNotification(t, tt.send, tt.key, true)
		})
	}
}

func TestAccountConfirmations(t *testing.T) {
	tests := []struct {
		name            string
		send            sender
		key             string
		confirm, cancel string
	}{
		{"DeleteAccountConfirmation", DeleteAccountConfirmation, "delete_account.confirm", weburl.AccountDeleteConfirm, weburl.AccountDeleteCancel},
		{"DisableAccountConfirmation", DisableAccountConfirmation, "disable_account.confirm", weburl.AccountDisableConfirm, weburl.AccountDisableCancel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkProcess(t, tt.send, tt.key, tt.confirm, tt.cancel)
		})
	}
}
