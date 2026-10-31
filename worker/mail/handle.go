package mail

import (
	"context"

	"github.com/cygy/ginamite/common/queue"
)

// handlers returns the senders of the emails, by type of message.
func (m *Mailer) handlers() map[string]func(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	return map[string]func(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc){
		queue.MessageMailRegistrationConfirmation: func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) {
			m.RegistrationConfirmation(payload)
		},
		queue.MessageMailRegistrationConfirmed: func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) {
			m.RegistrationConfirmed(payload)
		},
		queue.MessageMailRegistrationCancelled: func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) {
			m.RegistrationCancelled(payload)
		},
		queue.MessageMailRegistrationWelcome: func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) {
			m.RegistrationWelcome(payload)
		},
		queue.MessageMailForgotPasswordConfirmation:     m.ForgotPasswordConfirmation,
		queue.MessageMailForgotPasswordConfirmed:        m.ForgotPasswordConfirmed,
		queue.MessageMailForgotPasswordCancelled:        m.ForgotPasswordCancelled,
		queue.MessageMailUpdateEmailAddressConfirmation: m.UpdateEmailAddressConfirmation,
		queue.MessageMailUpdateEmailAddressConfirmed:    m.UpdateEmailAddressConfirmed,
		queue.MessageMailUpdateEmailAddressCancelled:    m.UpdateEmailAddressCancelled,
		queue.MessageMailUpdatePasswordConfirmation:     m.UpdatePasswordConfirmation,
		queue.MessageMailUpdatePasswordConfirmed:        m.UpdatePasswordConfirmed,
		queue.MessageMailUpdatePasswordCancelled:        m.UpdatePasswordCancelled,
		queue.MessageMailValidateNewEmailAddress:        m.ValidateNewEmailAddress,
		queue.MessageMailDeleteAccountConfirmation:      m.DeleteAccountConfirmation,
		queue.MessageMailDeleteAccountConfirmed:         m.DeleteAccountConfirmed,
		queue.MessageMailDeleteAccountCancelled:         m.DeleteAccountCancelled,
		queue.MessageMailDisableAccountConfirmation:     m.DisableAccountConfirmation,
		queue.MessageMailDisableAccountConfirmed:        m.DisableAccountConfirmed,
		queue.MessageMailDisableAccountCancelled:        m.DisableAccountCancelled,
		queue.MessageMailEnableAccountConfirmed:         m.EnableAccountConfirmed,
		queue.MessageMailNewLogin:                       m.NewLogin,
	}
}

// Handle sends the email of a message of the mail topic.
// It reports whether the type of the message is a built-in one.
func (m *Mailer) Handle(ctx context.Context, messageType string, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) bool {
	handler, ok := m.handlers()[messageType]
	if ok {
		handler(ctx, payload, isEmailAddressValid)
	}

	return ok
}
