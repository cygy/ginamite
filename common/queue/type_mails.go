package queue

// MailNotification : content of a payload
type MailNotification struct {
	EmailAddress string `json:"email"`
	Username     string `json:"username"`
	Locale       string `json:"locale"`
}

// MailProcessConfirmation : content of a payload
type MailProcessConfirmation struct {
	EmailAddress string `json:"email"`
	Username     string `json:"username"`
	Locale       string `json:"locale"`
	Data         string `json:"data"`
	ProcessID    string `json:"process_id"`
	ProcessCode  string `json:"process_code"`
}

// MailRegistrationConfirmation : content of a payload
type MailRegistrationConfirmation struct {
	EmailAddress     string `json:"email"`
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	Locale           string `json:"locale"`
	RegistrationCode string `json:"registration_code"`
}

// MailRegistrationWelcome : content of a payload
type MailRegistrationWelcome MailNotification

// MailRegistrationConfirmed : content of a payload
type MailRegistrationConfirmed MailNotification

// MailRegistrationCancelled : content of a payload
type MailRegistrationCancelled MailNotification

// MailForgotPasswordConfirmation : content of a payload
type MailForgotPasswordConfirmation MailProcessConfirmation

// MailForgotPasswordConfirmed : content of a payload
type MailForgotPasswordConfirmed MailNotification

// MailForgotPasswordCancelled : content of a payload
type MailForgotPasswordCancelled MailNotification

// MailUpdateEmailAddressConfirmation : content of a payload
type MailUpdateEmailAddressConfirmation MailProcessConfirmation

// MailUpdateEmailAddressConfirmed : content of a payload
type MailUpdateEmailAddressConfirmed MailNotification

// MailUpdateEmailAddressCancelled : content of a payload
type MailUpdateEmailAddressCancelled MailNotification

// MailUpdatePasswordConfirmation : content of a payload
type MailUpdatePasswordConfirmation MailProcessConfirmation

// MailUpdatePasswordConfirmed : content of a payload
type MailUpdatePasswordConfirmed MailNotification

// MailUpdatePasswordCancelled : content of a payload
type MailUpdatePasswordCancelled MailNotification

// MailValidateNewEmailAddress : content of a payload
type MailValidateNewEmailAddress MailProcessConfirmation

// MailDeleteAccountConfirmation : content of a payload
type MailDeleteAccountConfirmation MailProcessConfirmation

// MailDeleteAccountConfirmed : content of a payload
type MailDeleteAccountConfirmed MailNotification

// MailDeleteAccountCancelled : content of a payload
type MailDeleteAccountCancelled MailNotification

// MailDisableAccountConfirmation : content of a payload
type MailDisableAccountConfirmation MailProcessConfirmation

// MailDisableAccountConfirmed : content of a payload
type MailDisableAccountConfirmed MailNotification

// MailEnableAccountConfirmed : content of a payload
type MailEnableAccountConfirmed MailNotification

// MailDisableAccountCancelled : content of a payload
type MailDisableAccountCancelled MailNotification

// MailNewLogin : content of a payload
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
