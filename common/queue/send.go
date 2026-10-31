package queue

import "time"

// CreateTask sends a new task to the tasks topic, for the workers.
func (q *Queue) CreateTask(taskName string, payload any) {
	q.sendMessage(taskName, payload, TopicTasks)
}

// GetIPLocation asks the workers to retrieve the location of an IP address.
func (q *Queue) GetIPLocation(payload TaskIPLocation) {
	q.CreateTask(MessageTaskIPLocation, payload)
}

// DeleteUser asks the workers to delete the data of a user.
func (q *Queue) DeleteUser(payload TaskDeleteUser) {
	q.CreateTask(MessageTaskDeleteUser, payload)
}

// DisableUser asks the workers to disable the data of a user.
func (q *Queue) DisableUser(payload TaskDisableUser) {
	q.CreateTask(MessageTaskDisableUser, payload)
}

// EnableUser asks the workers to enable the data of a user.
func (q *Queue) EnableUser(payload TaskEnableUser) {
	q.CreateTask(MessageTaskEnableUser, payload)
}

// UpdateUserSocialNetworks asks the workers to update the social networks of a user.
func (q *Queue) UpdateUserSocialNetworks(payload TaskUpdateUserSocialNetworks) {
	q.CreateTask(MessageTaskUpdateUserSocialNetworks, payload)
}

// RegistrationDone informs the workers that a new user is registered.
func (q *Queue) RegistrationDone(payload TaskRegistrationDone) {
	q.CreateTask(MessageTaskRegistrationDone, payload)
}

// RegistrationValidated informs the workers that a new user is validated.
func (q *Queue) RegistrationValidated(payload TaskRegistrationValidated) {
	q.CreateTask(MessageTaskRegistrationValidated, payload)
}

// InvalidateCache sends a cache message to the cache topic.
func (q *Queue) InvalidateCache(message string, payload any) {
	q.sendMessage(message, payload, TopicCache)
}

// InvalidAuthToken invalidates the auth token tokenID until its expiration date.
func (q *Queue) InvalidAuthToken(tokenID string, expirationDate time.Time) {
	payload := CacheAuthToken{
		Tokens: map[string]time.Time{
			tokenID: expirationDate,
		},
	}

	q.InvalidateCache(MessageCacheInvalidAuthToken, payload)
}

// InvalidAuthTokens invalidates several auth tokens, given with their expiration date by token ID.
func (q *Queue) InvalidAuthTokens(tokens map[string]time.Time) {
	payload := CacheAuthToken{
		Tokens: tokens,
	}

	q.InvalidateCache(MessageCacheInvalidAuthToken, payload)
}

// NotifyUsersWithAbility notifies the users having the ability that an event occurs.
func (q *Queue) NotifyUsersWithAbility(ability, messageType string, payload any) {
	groupPayload := GroupNotification{
		Ability: ability,
		Payload: payload,
	}
	q.sendMessage(messageType, groupPayload, TopicGroupNotification)
}

// NotifyUser notifies a user that an event occurs.
func (q *Queue) NotifyUser(messageType string, payload any) {
	q.sendMessage(messageType, payload, TopicUserNotification)
}

// SendMail sends a mail to a user because an event occurs.
func (q *Queue) SendMail(messageType string, payload any) {
	q.sendMessage(messageType, payload, TopicMail)
}

// SendChromeNotification notifies a user, through Chrome, that an event occurs.
func (q *Queue) SendChromeNotification(messageType string, payload any) {
	q.sendMessage(messageType, payload, TopicChrome)
}

// SendSafariNotification notifies a user, through Safari, that an event occurs.
func (q *Queue) SendSafariNotification(messageType string, payload any) {
	q.sendMessage(messageType, payload, TopicSafari)
}

// SendFirefoxNotification notifies a user, through Firefox, that an event occurs.
func (q *Queue) SendFirefoxNotification(messageType string, payload any) {
	q.sendMessage(messageType, payload, TopicFirefox)
}

// SendIOSNotification notifies a user, through the iOS app, that an event occurs.
func (q *Queue) SendIOSNotification(messageType string, payload any) {
	q.sendMessage(messageType, payload, TopicIOS)
}

// SendAndroidNotification notifies a user, through the Android app, that an event occurs.
func (q *Queue) SendAndroidNotification(messageType string, payload any) {
	q.sendMessage(messageType, payload, TopicAndroid)
}
