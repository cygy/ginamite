package queue

// NotifyUsersWithAbility : notify the users with an ability that an event occurs.
func NotifyUsersWithAbility(ability, messageType string, payload interface{}) {
	groupPayload := GroupNotification{
		Ability: ability,
		Payload: payload,
	}
	sendMessage(messageType, groupPayload, TopicGroupNotification)
}

// NotifyUser : notify the user that an event occurs.
func NotifyUser(messageType string, payload interface{}) {
	sendMessage(messageType, payload, TopicUserNotification)
}

// SendMail : notify the user that an event occurs by email.
func SendMail(messageType string, payload interface{}) {
	sendMessage(messageType, payload, TopicMail)
}

// SendChromeNotification : notify the user that an event occurs by chrome.
func SendChromeNotification(messageType string, payload interface{}) {
	sendMessage(messageType, payload, TopicChrome)
}

// SendSafariNotification : notify the user that an event occurs by safari.
func SendSafariNotification(messageType string, payload interface{}) {
	sendMessage(messageType, payload, TopicSafari)
}

// SendFirefoxNotification : notify the user that an event occurs by firefox.
func SendFirefoxNotification(messageType string, payload interface{}) {
	sendMessage(messageType, payload, TopicFirefox)
}

// SendIOSNotification : notify the user that an event occurs by iOS.
func SendIOSNotification(messageType string, payload interface{}) {
	sendMessage(messageType, payload, TopicIOS)
}

// SendAndroidNotification : notify the user that an event occurs by android.
func SendAndroidNotification(messageType string, payload interface{}) {
	sendMessage(messageType, payload, TopicAndroid)
}
