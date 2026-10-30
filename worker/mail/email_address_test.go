package mail

import (
	"testing"

	"github.com/cygy/ginamite/common/weburl"
)

func TestUpdateEmailAddressConfirmation(t *testing.T) {
	checkProcess(t, UpdateEmailAddressConfirmation, "update_email_address.confirm", weburl.AccountUpdateEmailAddressConfirm, weburl.AccountUpdateEmailAddressCancel)
}

func TestUpdateEmailAddressConfirmed(t *testing.T) {
	checkNotification(t, UpdateEmailAddressConfirmed, "update_email_address.confirmed", true)
}

func TestUpdateEmailAddressCancelled(t *testing.T) {
	checkNotification(t, UpdateEmailAddressCancelled, "update_email_address.cancelled", true)
}

func TestValidateNewEmailAddress(t *testing.T) {
	checkProcess(t, ValidateNewEmailAddress, "validate_new_email_address", weburl.AccountValidateNewEmailAddress, "")
}
