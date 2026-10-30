package ginamite

import "github.com/cygy/ginamite/common/notifications"

// ApplicationConfiguration is the additional and optional configuration of the application.
type ApplicationConfiguration struct {
	FileRequired   bool                  // If true, the application configuration file is mandatory.
	InitializeFunc func(filePath string) // Called with the path of the application configuration file.
}

// AddSupportedNotification adds a supported notification type, enabled by default or not.
func (a ApplicationConfiguration) AddSupportedNotification(notification string, isDefault bool) {
	notifications.AddNotification(notification, isDefault)
}

// DisableNotification disables a supported notification and reports whether it exists.
func (a ApplicationConfiguration) DisableNotification(notification string) bool {
	return notifications.DisableNotification(notification)
}

// EnableNotification enables a supported notification and reports whether it exists.
func (a ApplicationConfiguration) EnableNotification(notification string) bool {
	return notifications.EnableNotification(notification)
}

// SetNotificationDefault sets a supported notification as enabled by default or not, and reports whether it exists.
func (a ApplicationConfiguration) SetNotificationDefault(notification string, isDefault bool) bool {
	return notifications.SetNotificationDefault(notification, isDefault)
}
