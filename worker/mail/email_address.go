package mail

import (
	"context"

	"github.com/cygy/ginamite/common/weburl"
)

var (
	updateEmailAddress      = process{"update_email_address.confirm", weburl.AccountUpdateEmailAddressConfirm, weburl.AccountUpdateEmailAddressCancel}
	validateNewEmailAddress = process{key: "validate_new_email_address", confirm: weburl.AccountValidateNewEmailAddress}
)

// UpdateEmailAddressConfirmation sends an email asking to confirm, or to cancel, the update of the email address.
func UpdateEmailAddressConfirmation(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendProcess(ctx, payload, updateEmailAddress, isEmailAddressValid)
}

// UpdateEmailAddressConfirmed sends an email to notify the update of the email address is confirmed.
func UpdateEmailAddressConfirmed(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendNotification(ctx, payload, "update_email_address.confirmed", isEmailAddressValid)
}

// UpdateEmailAddressCancelled sends an email to notify the update of the email address is cancelled.
func UpdateEmailAddressCancelled(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendNotification(ctx, payload, "update_email_address.cancelled", isEmailAddressValid)
}

// ValidateNewEmailAddress sends an email asking to validate the new email address.
func ValidateNewEmailAddress(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	sendProcess(ctx, payload, validateNewEmailAddress, isEmailAddressValid)
}
