package mail

import (
	"context"

	"github.com/cygy/ginamite/common/weburl"
)

var (
	forgotPassword = process{"forgot_password.confirm", weburl.AccountForgotPasswordConfirm, weburl.AccountForgotPasswordCancel}
	updatePassword = process{"update_password.confirm", weburl.AccountUpdatePasswordConfirm, weburl.AccountUpdatePasswordCancel}
)

// ForgotPasswordConfirmation sends an email asking to confirm, or to cancel, the reset of the password.
func ForgotPasswordConfirmation(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendProcess(ctx, payload, forgotPassword, isEmailAddressValid)
}

// ForgotPasswordConfirmed sends an email to notify the reset of the password is confirmed.
func ForgotPasswordConfirmed(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendNotification(ctx, payload, "forgot_password.confirmed", isEmailAddressValid)
}

// ForgotPasswordCancelled sends an email to notify the reset of the password is cancelled.
func ForgotPasswordCancelled(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendNotification(ctx, payload, "forgot_password.cancelled", isEmailAddressValid)
}

// UpdatePasswordConfirmation sends an email asking to confirm, or to cancel, the update of the password.
func UpdatePasswordConfirmation(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendProcess(ctx, payload, updatePassword, isEmailAddressValid)
}

// UpdatePasswordConfirmed sends an email to notify the update of the password is confirmed.
func UpdatePasswordConfirmed(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendNotification(ctx, payload, "update_password.confirmed", isEmailAddressValid)
}

// UpdatePasswordCancelled sends an email to notify the update of the password is cancelled.
func UpdatePasswordCancelled(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendNotification(ctx, payload, "update_password.cancelled", isEmailAddressValid)
}
