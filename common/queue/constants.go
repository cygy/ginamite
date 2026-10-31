// Package queue sends and decodes the messages exchanged through Kafka: tasks for
// the workers, cache invalidations, mails and notifications.
package queue

// Topics returns all the topics of the messages.
func Topics() []string {
	return []string{TopicTasks, TopicTaskLocation, TopicTaskUser, TopicTaskRegistration, TopicCache, TopicMail, TopicUserNotification, TopicGroupNotification, TopicIOS, TopicAndroid, TopicSafari, TopicFirefox, TopicChrome}
}

// Topics of the messages.
const (
	// TopicTasks is the name of the topic for the custom tasks, see Queue.CreateTask. It also still carries the built-in
	// tasks sent before they had their own topics, which the workers keep handling.
	TopicTasks = "task"

	// TopicTaskLocation is the name of the topic for the location of the IP addresses: slow tasks, which depend on an
	// external service.
	TopicTaskLocation = "task_location"

	// TopicTaskUser is the name of the topic for the tasks on the data of the users (deleting, disabling, enabling,
	// social networks), ordered by user.
	TopicTaskUser = "task_user"

	// TopicTaskRegistration is the name of the topic for the events of the registration of the users.
	TopicTaskRegistration = "task_registration"

	// TopicCache is the name of the topic for the various caches.
	TopicCache = "cache"

	// TopicMail is the name of the topic for mail to send.
	TopicMail = "mail"

	// TopicUserNotification is the name of the topic for the user notifications.
	TopicUserNotification = "user_notification"

	// TopicGroupNotification is the name of the topic for the group notifications.
	TopicGroupNotification = "group_notification"

	// TopicIOS is the name of the topic for the ios notifications.
	TopicIOS = "ios"

	// TopicAndroid is the name of the topic for the android notifications.
	TopicAndroid = "android"

	// TopicSafari is the name of the topic for the safari notifications.
	TopicSafari = "safari"

	// TopicFirefox is the name of the topic for the firefox notifications.
	TopicFirefox = "firefox"

	// TopicChrome is the name of the topic for the chrome notifications.
	TopicChrome = "chrome"
)

// Types of the cache messages.
const (
	// MessageCacheInvalidAuthToken is the type of the cache message invalidating auth tokens.
	MessageCacheInvalidAuthToken = "invalid_auth_token"
)

// Types of the task messages.
const (
	// MessageTaskIPLocation is the type of the task which gets the location of an IP address.
	MessageTaskIPLocation = "ip_location"

	// MessageTaskDeleteUser is the type of the task which deletes the data of a user.
	MessageTaskDeleteUser = "delete_user"

	// MessageTaskDisableUser is the type of the task which disables the data of a user.
	MessageTaskDisableUser = "disable_user"

	// MessageTaskEnableUser is the type of the task which enables the data of a user.
	MessageTaskEnableUser = "enable_user"

	// MessageTaskUpdateUserSocialNetworks is the type of the task which updates the social networks of a user.
	MessageTaskUpdateUserSocialNetworks = "update_user_social_networks"

	// MessageTaskRegistrationDone is the type of the task which informs that a new user is registered.
	MessageTaskRegistrationDone = "registration_done"

	// MessageTaskRegistrationValidated is the type of the task which informs that a new user is validated.
	MessageTaskRegistrationValidated = "registration_validated"
)

// Types of the mail messages.
const (
	// MessageMailRegistrationConfirmation is the type of the mail sent when the registration must be confirmed.
	MessageMailRegistrationConfirmation = "registration_confirmation"

	// MessageMailRegistrationConfirmed is the type of the mail sent when the registration is confirmed.
	MessageMailRegistrationConfirmed = "registration_confirmed"

	// MessageMailRegistrationCancelled is the type of the mail sent when the registration is cancelled.
	MessageMailRegistrationCancelled = "registration_cancelled"

	// MessageMailRegistrationWelcome is the type of the mail sent when a new user is welcomed.
	MessageMailRegistrationWelcome = "registration_welcome"

	// MessageMailForgotPasswordConfirmation is the type of the mail sent when the password update must be confirmed.
	MessageMailForgotPasswordConfirmation = "forgot_password_confirmation"

	// MessageMailForgotPasswordConfirmed is the type of the mail sent when the password update is confirmed.
	MessageMailForgotPasswordConfirmed = "forgot_password_confirmed"

	// MessageMailForgotPasswordCancelled is the type of the mail sent when the password update is cancelled.
	MessageMailForgotPasswordCancelled = "forgot_password_cancelled"

	// MessageMailUpdateEmailAddressConfirmation is the type of the mail sent when the email address update must be confirmed.
	MessageMailUpdateEmailAddressConfirmation = "update_email_confirmation"

	// MessageMailUpdateEmailAddressConfirmed is the type of the mail sent when the email address update is confirmed.
	MessageMailUpdateEmailAddressConfirmed = "update_email_confirmed"

	// MessageMailUpdateEmailAddressCancelled is the type of the mail sent when the email address update is cancelled.
	MessageMailUpdateEmailAddressCancelled = "update_email_cancelled"

	// MessageMailUpdatePasswordConfirmation is the type of the mail sent when the password update must be confirmed.
	MessageMailUpdatePasswordConfirmation = "update_password_confirmation"

	// MessageMailUpdatePasswordConfirmed is the type of the mail sent when the password update is confirmed.
	MessageMailUpdatePasswordConfirmed = "update_password_confirmed"

	// MessageMailUpdatePasswordCancelled is the type of the mail sent when the password update is cancelled.
	MessageMailUpdatePasswordCancelled = "update_password_cancelled"

	// MessageMailValidateNewEmailAddress is the type of the mail sent when the new email address must be validated.
	MessageMailValidateNewEmailAddress = "validate_new_email"

	// MessageMailDeleteAccountConfirmation is the type of the mail sent when the account deletion must be confirmed.
	MessageMailDeleteAccountConfirmation = "delete_account_confirmation"

	// MessageMailDeleteAccountConfirmed is the type of the mail sent when the account deletion is confirmed.
	MessageMailDeleteAccountConfirmed = "delete_account_confirmed"

	// MessageMailDeleteAccountCancelled is the type of the mail sent when the account deletion is cancelled.
	MessageMailDeleteAccountCancelled = "delete_account_cancelled"

	// MessageMailDisableAccountConfirmation is the type of the mail sent when the account disabling must be confirmed.
	MessageMailDisableAccountConfirmation = "disable_account_confirmation"

	// MessageMailDisableAccountConfirmed is the type of the mail sent when the account disabling is confirmed.
	MessageMailDisableAccountConfirmed = "disable_account_confirmed"

	// MessageMailEnableAccountConfirmed is the type of the mail sent when the account enabling is confirmed.
	MessageMailEnableAccountConfirmed = "enable_account_confirmed"

	// MessageMailDisableAccountCancelled is the type of the mail sent when the account disabling is cancelled.
	MessageMailDisableAccountCancelled = "disable_account_cancelled"

	// MessageMailNewLogin is the type of the mail sent when a new login occurs.
	MessageMailNewLogin = "new_login"
)
