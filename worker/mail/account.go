package mail

import (
	"context"

	"github.com/cygy/ginamite/common/weburl"
)

var (
	deleteAccount  = process{"delete_account.confirm", weburl.AccountDeleteConfirm, weburl.AccountDeleteCancel}
	disableAccount = process{"disable_account.confirm", weburl.AccountDisableConfirm, weburl.AccountDisableCancel}
)

// DeleteAccountConfirmation sends an email asking to confirm, or to cancel, the deletion of the account.
func (m *Mailer) DeleteAccountConfirmation(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) error {
	return m.sendProcess(ctx, payload, deleteAccount, isEmailAddressValid)
}

// DeleteAccountConfirmed sends an email to notify the deletion of the account is confirmed.
func (m *Mailer) DeleteAccountConfirmed(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) error {
	return m.sendNotification(ctx, payload, "delete_account.confirmed", isEmailAddressValid)
}

// DeleteAccountCancelled sends an email to notify the deletion of the account is cancelled.
func (m *Mailer) DeleteAccountCancelled(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) error {
	return m.sendNotification(ctx, payload, "delete_account.cancelled", isEmailAddressValid)
}

// DisableAccountConfirmation sends an email asking to confirm, or to cancel, the deactivation of the account.
func (m *Mailer) DisableAccountConfirmation(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) error {
	return m.sendProcess(ctx, payload, disableAccount, isEmailAddressValid)
}

// DisableAccountConfirmed sends an email to notify the deactivation of the account is confirmed.
func (m *Mailer) DisableAccountConfirmed(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) error {
	return m.sendNotification(ctx, payload, "disable_account.confirmed", isEmailAddressValid)
}

// DisableAccountCancelled sends an email to notify the deactivation of the account is cancelled.
func (m *Mailer) DisableAccountCancelled(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) error {
	return m.sendNotification(ctx, payload, "disable_account.cancelled", isEmailAddressValid)
}

// EnableAccountConfirmed sends an email to notify the account is enabled again.
func (m *Mailer) EnableAccountConfirmed(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) error {
	return m.sendNotification(ctx, payload, "enable_account.confirmed", isEmailAddressValid)
}
