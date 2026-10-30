// Package notifications manages the types of notifications a user can subscribe to, and the targets receiving them.
package notifications

import (
	"context"
	"crypto/sha256"
	"fmt"
)

// respresents a notification that can be subscribed by a user.
type notificationType struct {
	Name    string
	Default bool
	Enabled bool
}

// Target is a destination of a notification sent to a user.
type Target struct {
	Target string // a value defined into the file 'constants_targets.go'
	Value  string // an email address, a token, etc...
}

// GetNotificationTargetsByUserAndTypeFunc returns the targets of a user for a type of notification.
type GetNotificationTargetsByUserAndTypeFunc func(ctx context.Context, userID, notificationType string) ([]Target, error)

// all the notifications that can be subscribed by a user.
var supportedNotifications = []notificationType{
	{
		Name:    TypeNewLogin,
		Default: true,
		Enabled: true,
	},
	{
		Name:    TypeCommercialPartners,
		Default: false,
		Enabled: true,
	},
}

// IsNotificationSupported reports whether the notification is supported and enabled.
func IsNotificationSupported(notification string) bool {
	for _, supportedNotification := range supportedNotifications {
		if notification == supportedNotification.Name {
			return supportedNotification.Enabled
		}
	}

	return false
}

// AddNotification adds a notification type to the supported list.
// An existing notification is enabled again and its default flag is updated.
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

// EnableNotification enables a notification and reports whether it is in the supported list.
func EnableNotification(notification string) bool {
	return enableOrDisableNotification(notification, true)
}

// DisableNotification disables a notification and reports whether it is in the supported list.
func DisableNotification(notification string) bool {
	return enableOrDisableNotification(notification, false)
}

// SetNotificationDefault sets whether a notification is subscribed by default
// and reports whether it is in the supported list.
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

// SupportedNotifications returns the supported and enabled notifications.
func SupportedNotifications() []string {
	notifications := []string{}
	for _, supportedNotification := range supportedNotifications {
		if supportedNotification.Enabled {
			notifications = append(notifications, supportedNotification.Name)
		}
	}

	return notifications
}

// DefaultNotifications returns the supported and enabled default notifications.
func DefaultNotifications() []string {
	notifications := []string{}
	for _, supportedNotification := range supportedNotifications {
		if supportedNotification.Enabled && supportedNotification.Default {
			notifications = append(notifications, supportedNotification.Name)
		}
	}

	return notifications
}

// GenerateUnsubscribeKey generates a key "userID.notificationType.hash" that lets a
// user unsubscribe from a notification, from an email, without being logged in.
// The hash depends on the private key of the user, so it can not be forged.
func GenerateUnsubscribeKey(userID, privateKey, notificationType string) string {
	sha256 := sha256.New()
	sha256.Write([]byte(userID + notificationType + privateKey))
	encodedKey := sha256.Sum(nil)

	return fmt.Sprintf("%s.%s.%s", userID, notificationType, fmt.Sprintf("%x", encodedKey))
}

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
