package mail

import (
	"context"

	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// RegistrationWelcome sends an email to welcome the user after the registration.
func RegistrationWelcome(payload []byte) {
	sendNotification(context.Background(), payload, "registration.welcome", alwaysValid)
}

// RegistrationConfirmation sends an email asking to confirm, or to cancel, the registration.
func RegistrationConfirmation(payload []byte) {
	p := queue.MailRegistrationConfirmation{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	deliver(content{
		key:        "registration.confirm",
		template:   TemplateConfirmationAndCancel,
		recipient:  p.EmailAddress,
		username:   p.Username,
		locale:     p.Locale,
		confirmURL: processURL(weburl.AccountRegistrationConfirm, p.Locale, p.UserID, p.RegistrationCode),
		cancelURL:  processURL(weburl.AccountRegistrationCancel, p.Locale, p.UserID, p.RegistrationCode),
	}.message())
}

// RegistrationConfirmed sends an email to notify the registration is confirmed.
func RegistrationConfirmed(payload []byte) {
	sendNotification(context.Background(), payload, "registration.confirmed", alwaysValid)
}

// RegistrationCancelled sends an email to notify the registration is cancelled.
func RegistrationCancelled(payload []byte) {
	sendNotification(context.Background(), payload, "registration.cancelled", alwaysValid)
}
