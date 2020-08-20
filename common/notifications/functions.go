package notifications

import (
	"crypto/sha256"
	"fmt"
)

// IsNotificationSupported : returns true if the notification is a supported notification.
func IsNotificationSupported(notification string) bool {
	for _, supportedNotification := range supportedNotifications {
		if notification == supportedNotification.Name {
			return supportedNotification.Enabled
		}
	}

	return false
}

// AddNotification : add a notification type to the supported list.
func AddNotification(notification string, isDefault bool) {
	// If the notification is supported already, it is enabled and set as default if needed.
	for i, supportedNotification := range supportedNotifications {
		if notification == supportedNotification.Name {
			supportedNotification.Enabled = true
			supportedNotification.Default = isDefault
			supportedNotifications[i] = supportedNotification
			return
		}
	}

	supportedNotifications = append(supportedNotifications, notificationType{
		Name:    notification,
		Default: isDefault,
		Enabled: true,
	})
}

// EnableNotification : enables a notification if it is in the supported list.
func EnableNotification(notification string) bool {
	return enableOrDisableNotification(notification, true)
}

// DisableNotification : disables a notification if it is in the supported list.
func DisableNotification(notification string) bool {
	return enableOrDisableNotification(notification, false)
}

// SetNotificationDefault : sets a notification as default or not.
func SetNotificationDefault(notification string, isDefault bool) bool {
	for i, supportedNotification := range supportedNotifications {
		if notification == supportedNotification.Name {
			supportedNotification.Default = isDefault
			supportedNotifications[i] = supportedNotification
			return true
		}
	}

	return false
}

// SupportedNotifications : returns the supported and enabled notifications.
func SupportedNotifications() []string {
	notifications := []string{}
	for _, supportedNotification := range supportedNotifications {
		if supportedNotification.Enabled {
			notifications = append(notifications, supportedNotification.Name)
		}
	}

	return notifications
}

// DefaultNotifications : returns the supported and enabled default notifications.
func DefaultNotifications() []string {
	notifications := []string{}
	for _, supportedNotification := range supportedNotifications {
		if supportedNotification.Enabled && supportedNotification.Default {
			notifications = append(notifications, supportedNotification.Name)
		}
	}

	return notifications
}

// GenerateUnsubscribeKey : generates an unsubscribe key. This key is using to unsubscribe a notification by email without authorization.
func GenerateUnsubscribeKey(userID, privateKey, notificationType string) string {
	sha256 := sha256.New()
	sha256.Write([]byte(userID + notificationType + privateKey))
	encodedKey := sha256.Sum(nil)

	return fmt.Sprintf("%s.%s.%s", userID, notificationType, fmt.Sprintf("%x", encodedKey))
}
