package queue

import "time"

// CacheAuthToken is the payload of the message MessageCacheInvalidAuthToken: the expiration date of each invalid auth token, by token ID.
type CacheAuthToken struct {
	Tokens map[string]time.Time `json:"tokens"`
}

// TaskIPLocation is the payload of the message MessageTaskIPLocation.
type TaskIPLocation struct {
	IPAddress string `json:"ip"`
	TokenID   string `json:"token_id"`
}

// TaskDeleteUser is the payload of the message MessageTaskDeleteUser.
type TaskDeleteUser struct {
	UserID string `json:"user_id"`
}

// TaskDisableUser is the payload of the message MessageTaskDisableUser.
type TaskDisableUser struct {
	UserID string `json:"user_id"`
}

// TaskEnableUser is the payload of the message MessageTaskEnableUser.
type TaskEnableUser struct {
	UserID string `json:"user_id"`
}

// TaskUpdateUserSocialNetworks is the payload of the message MessageTaskUpdateUserSocialNetworks.
type TaskUpdateUserSocialNetworks struct {
	UserID string `json:"user_id"`
}

// TaskRegistrationDone is the payload of the message MessageTaskRegistrationDone.
type TaskRegistrationDone struct {
	UserID string `json:"user_id"`
}

// TaskRegistrationValidated is the payload of the message MessageTaskRegistrationValidated.
type TaskRegistrationValidated struct {
	UserID string `json:"user_id"`
}

// MailNotification is the base payload of the mails sent to a user, identified by their email address.
type MailNotification struct {
	EmailAddress string `json:"email"`
	Username     string `json:"username"`
	Locale       string `json:"locale"`
}

// MailProcessConfirmation is the base payload of the mails asking a user to confirm a process.
type MailProcessConfirmation struct {
	EmailAddress string `json:"email"`
	Username     string `json:"username"`
	Locale       string `json:"locale"`
	Data         string `json:"data"`
	ProcessID    string `json:"process_id"`
	ProcessCode  string `json:"process_code"`
}

// MailRegistrationConfirmation is the payload of the message MessageMailRegistrationConfirmation.
type MailRegistrationConfirmation struct {
	EmailAddress     string `json:"email"`
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	Locale           string `json:"locale"`
	RegistrationCode string `json:"registration_code"`
}

// MailRegistrationWelcome is the payload of the message MessageMailRegistrationWelcome.
type MailRegistrationWelcome MailNotification

// MailRegistrationConfirmed is the payload of the message MessageMailRegistrationConfirmed.
type MailRegistrationConfirmed MailNotification

// MailRegistrationCancelled is the payload of the message MessageMailRegistrationCancelled.
type MailRegistrationCancelled MailNotification

// MailForgotPasswordConfirmation is the payload of the message MessageMailForgotPasswordConfirmation.
type MailForgotPasswordConfirmation MailProcessConfirmation

// MailForgotPasswordConfirmed is the payload of the message MessageMailForgotPasswordConfirmed.
type MailForgotPasswordConfirmed MailNotification

// MailForgotPasswordCancelled is the payload of the message MessageMailForgotPasswordCancelled.
type MailForgotPasswordCancelled MailNotification

// MailUpdateEmailAddressConfirmation is the payload of the message MessageMailUpdateEmailAddressConfirmation.
type MailUpdateEmailAddressConfirmation MailProcessConfirmation

// MailUpdateEmailAddressConfirmed is the payload of the message MessageMailUpdateEmailAddressConfirmed.
type MailUpdateEmailAddressConfirmed MailNotification

// MailUpdateEmailAddressCancelled is the payload of the message MessageMailUpdateEmailAddressCancelled.
type MailUpdateEmailAddressCancelled MailNotification

// MailUpdatePasswordConfirmation is the payload of the message MessageMailUpdatePasswordConfirmation.
type MailUpdatePasswordConfirmation MailProcessConfirmation

// MailUpdatePasswordConfirmed is the payload of the message MessageMailUpdatePasswordConfirmed.
type MailUpdatePasswordConfirmed MailNotification

// MailUpdatePasswordCancelled is the payload of the message MessageMailUpdatePasswordCancelled.
type MailUpdatePasswordCancelled MailNotification

// MailValidateNewEmailAddress is the payload of the message MessageMailValidateNewEmailAddress.
type MailValidateNewEmailAddress MailProcessConfirmation

// MailDeleteAccountConfirmation is the payload of the message MessageMailDeleteAccountConfirmation.
type MailDeleteAccountConfirmation MailProcessConfirmation

// MailDeleteAccountConfirmed is the payload of the message MessageMailDeleteAccountConfirmed.
type MailDeleteAccountConfirmed MailNotification

// MailDeleteAccountCancelled is the payload of the message MessageMailDeleteAccountCancelled.
type MailDeleteAccountCancelled MailNotification

// MailDisableAccountConfirmation is the payload of the message MessageMailDisableAccountConfirmation.
type MailDisableAccountConfirmation MailProcessConfirmation

// MailDisableAccountConfirmed is the payload of the message MessageMailDisableAccountConfirmed.
type MailDisableAccountConfirmed MailNotification

// MailEnableAccountConfirmed is the payload of the message MessageMailEnableAccountConfirmed.
type MailEnableAccountConfirmed MailNotification

// MailDisableAccountCancelled is the payload of the message MessageMailDisableAccountCancelled.
type MailDisableAccountCancelled MailNotification

// MailNewLogin is the payload of the message MessageMailNewLogin.
type MailNewLogin struct {
	EmailAddress   string `json:"email"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	UnsubscribeKey string `json:"unsubscribe_key"`
	IPAddress      string `json:"ip"`
	Source         string `json:"source"`
	Device         string `json:"device"`
	Locale         string `json:"locale"`
	TokenID        string `json:"token_id"`
	TokenKey       string `json:"token_key"`
}

// UserNotificationNewLogin is the payload of the user notification sent on a new login.
type UserNotificationNewLogin struct {
	EmailAddress   string `json:"email"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	UnsubscribeKey string `json:"unsubscribe_key"`
	IPAddress      string `json:"ip"`
	Source         string `json:"source"`
	Device         string `json:"device"`
	Locale         string `json:"locale"`
	TokenID        string `json:"token_id"`
	TokenKey       string `json:"token_key"`
}

// GroupNotification is the payload of the notifications sent to the users having an ability.
type GroupNotification struct {
	Ability string `json:"ability"`
	Payload any    `json:"payload"`
}
