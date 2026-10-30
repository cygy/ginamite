package mail

import (
	"context"

	"github.com/cygy/ginamite/common/queue"
)

// handlers are the senders of the emails, by type of message.
var handlers = map[string]func(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc){
	queue.MessageMailRegistrationConfirmation: func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) {
		RegistrationConfirmation(payload)
	},
	queue.MessageMailRegistrationConfirmed: func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) {
		RegistrationConfirmed(payload)
	},
	queue.MessageMailRegistrationCancelled: func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) {
		RegistrationCancelled(payload)
	},
	queue.MessageMailRegistrationWelcome: func(_ context.Context, payload []byte, _ IsEmailAddressValidFunc) {
		RegistrationWelcome(payload)
	},
	queue.MessageMailForgotPasswordConfirmation:     ForgotPasswordConfirmation,
	queue.MessageMailForgotPasswordConfirmed:        ForgotPasswordConfirmed,
	queue.MessageMailForgotPasswordCancelled:        ForgotPasswordCancelled,
	queue.MessageMailUpdateEmailAddressConfirmation: UpdateEmailAddressConfirmation,
	queue.MessageMailUpdateEmailAddressConfirmed:    UpdateEmailAddressConfirmed,
	queue.MessageMailUpdateEmailAddressCancelled:    UpdateEmailAddressCancelled,
	queue.MessageMailUpdatePasswordConfirmation:     UpdatePasswordConfirmation,
	queue.MessageMailUpdatePasswordConfirmed:        UpdatePasswordConfirmed,
	queue.MessageMailUpdatePasswordCancelled:        UpdatePasswordCancelled,
	queue.MessageMailValidateNewEmailAddress:        ValidateNewEmailAddress,
	queue.MessageMailDeleteAccountConfirmation:      DeleteAccountConfirmation,
	queue.MessageMailDeleteAccountConfirmed:         DeleteAccountConfirmed,
	queue.MessageMailDeleteAccountCancelled:         DeleteAccountCancelled,
	queue.MessageMailDisableAccountConfirmation:     DisableAccountConfirmation,
	queue.MessageMailDisableAccountConfirmed:        DisableAccountConfirmed,
	queue.MessageMailDisableAccountCancelled:        DisableAccountCancelled,
	queue.MessageMailEnableAccountConfirmed:         EnableAccountConfirmed,
	queue.MessageMailNewLogin:                       NewLogin,
}

// Handle sends the email of a message of the mail topic.
// It reports whether the type of the message is a built-in one.
func Handle(ctx context.Context, messageType string, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) bool {
	handler, ok := handlers[messageType]
	if ok {
		handler(ctx, payload, isEmailAddressValid)
	}

	return ok
}
