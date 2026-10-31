package mail

import (
	"context"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// RegistrationWelcome sends an email to welcome the user after the registration.
func (m *Mailer) RegistrationWelcome(payload []byte) error {
	return m.sendNotification(context.Background(), payload, "registration.welcome", alwaysValid)
}

// RegistrationConfirmation sends an email asking to confirm, or to cancel, the registration.
func (m *Mailer) RegistrationConfirmation(payload []byte) error {
	p := queue.MailRegistrationConfirmation{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	return m.deliver(m.message(content{
		key:        "registration.confirm",
		template:   TemplateConfirmationAndCancel,
		recipient:  p.EmailAddress,
		username:   p.Username,
		locale:     p.Locale,
		confirmURL: m.processURL(weburl.AccountRegistrationConfirm, p.Locale, p.UserID, p.RegistrationCode),
		cancelURL:  m.processURL(weburl.AccountRegistrationCancel, p.Locale, p.UserID, p.RegistrationCode),
	}))
}

// RegistrationConfirmed sends an email to notify the registration is confirmed.
func (m *Mailer) RegistrationConfirmed(payload []byte) error {
	return m.sendNotification(context.Background(), payload, "registration.confirmed", alwaysValid)
}

// RegistrationCancelled sends an email to notify the registration is cancelled.
func (m *Mailer) RegistrationCancelled(payload []byte) error {
	return m.sendNotification(context.Background(), payload, "registration.cancelled", alwaysValid)
}
