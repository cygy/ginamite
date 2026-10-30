package queue

import "time"

// CreateTask sends a new task to the tasks topic, for the workers.
func CreateTask(taskName string, payload any) {
	sendMessage(taskName, payload, TopicTasks)
}

// GetIPLocation asks the workers to retrieve the location of an IP address.
func GetIPLocation(payload TaskIPLocation) {
	CreateTask(MessageTaskIPLocation, payload)
}

// DeleteUser asks the workers to delete the data of a user.
func DeleteUser(payload TaskDeleteUser) {
	CreateTask(MessageTaskDeleteUser, payload)
}

// DisableUser asks the workers to disable the data of a user.
func DisableUser(payload TaskDisableUser) {
	CreateTask(MessageTaskDisableUser, payload)
}

// EnableUser asks the workers to enable the data of a user.
func EnableUser(payload TaskEnableUser) {
	CreateTask(MessageTaskEnableUser, payload)
}

// UpdateUserSocialNetworks asks the workers to update the social networks of a user.
func UpdateUserSocialNetworks(payload TaskUpdateUserSocialNetworks) {
	CreateTask(MessageTaskUpdateUserSocialNetworks, payload)
}

// RegistrationDone informs the workers that a new user is registered.
func RegistrationDone(payload TaskRegistrationDone) {
	CreateTask(MessageTaskRegistrationDone, payload)
}

// RegistrationValidated informs the workers that a new user is validated.
func RegistrationValidated(payload TaskRegistrationValidated) {
	CreateTask(MessageTaskRegistrationValidated, payload)
}

// InvalidateCache sends a cache message to the cache topic.
func InvalidateCache(message string, payload any) {
	sendMessage(message, payload, TopicCache)
}

// InvalidAuthToken invalidates the auth token tokenID until its expiration date.
func InvalidAuthToken(tokenID string, expirationDate time.Time) {
	payload := CacheAuthToken{
		Tokens: map[string]time.Time{
			tokenID: expirationDate,
		},
	}

	InvalidateCache(MessageCacheInvalidAuthToken, payload)
}

// InvalidAuthTokens invalidates several auth tokens, given with their expiration date by token ID.
func InvalidAuthTokens(tokens map[string]time.Time) {
	payload := CacheAuthToken{
		Tokens: tokens,
	}

	InvalidateCache(MessageCacheInvalidAuthToken, payload)
}

// NotifyUsersWithAbility notifies the users having the ability that an event occurs.
func NotifyUsersWithAbility(ability, messageType string, payload any) {
	groupPayload := GroupNotification{
		Ability: ability,
		Payload: payload,
	}
	sendMessage(messageType, groupPayload, TopicGroupNotification)
}

// NotifyUser notifies a user that an event occurs.
func NotifyUser(messageType string, payload any) {
	sendMessage(messageType, payload, TopicUserNotification)
}

// SendMail sends a mail to a user because an event occurs.
func SendMail(messageType string, payload any) {
	sendMessage(messageType, payload, TopicMail)
}

// SendChromeNotification notifies a user, through Chrome, that an event occurs.
func SendChromeNotification(messageType string, payload any) {
	sendMessage(messageType, payload, TopicChrome)
}

// SendSafariNotification notifies a user, through Safari, that an event occurs.
func SendSafariNotification(messageType string, payload any) {
	sendMessage(messageType, payload, TopicSafari)
}

// SendFirefoxNotification notifies a user, through Firefox, that an event occurs.
func SendFirefoxNotification(messageType string, payload any) {
	sendMessage(messageType, payload, TopicFirefox)
}

// SendIOSNotification notifies a user, through the iOS app, that an event occurs.
func SendIOSNotification(messageType string, payload any) {
	sendMessage(messageType, payload, TopicIOS)
}

// SendAndroidNotification notifies a user, through the Android app, that an event occurs.
func SendAndroidNotification(messageType string, payload any) {
	sendMessage(messageType, payload, TopicAndroid)
}
