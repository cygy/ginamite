package notifications

func enableOrDisableNotification(notification string, enabled bool) bool {
	for i, supportedNotification := range supportedNotifications {
		if notification == supportedNotification.Name {
			supportedNotification.Enabled = enabled
			supportedNotifications[i] = supportedNotification
			return true
		}
	}

	return false
}
