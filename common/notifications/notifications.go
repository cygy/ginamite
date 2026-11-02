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

// Registry is the list of the notifications a user can subscribe to. Create it with NewRegistry, and configure it
// before the servers start: it is not safe for concurrent use.
type Registry struct {
	supported []notificationType
}

// NewRegistry returns a registry with the built-in notifications.
func NewRegistry() *Registry {
	return &Registry{supported: []notificationType{
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
	}}
}

// IsSupported reports whether the notification is supported and enabled.
func (r *Registry) IsSupported(notification string) bool {
	for _, supportedNotification := range r.supported {
		if notification == supportedNotification.Name {
			return supportedNotification.Enabled
		}
	}

	return false
}

// Add adds a notification type to the supported list.
// An existing notification is enabled again and its default flag is updated.
func (r *Registry) Add(notification string, isDefault bool) {
	// If the notification is supported already, it is enabled and set as default if needed.
	for i, supportedNotification := range r.supported {
		if notification == supportedNotification.Name {
			supportedNotification.Enabled = true
			supportedNotification.Default = isDefault
			r.supported[i] = supportedNotification
			return
		}
	}

	r.supported = append(r.supported, notificationType{
		Name:    notification,
		Default: isDefault,
		Enabled: true,
	})
}

// Enable enables a notification and reports whether it is in the supported list.
func (r *Registry) Enable(notification string) bool {
	return r.enableOrDisable(notification, true)
}

// Disable disables a notification and reports whether it is in the supported list.
func (r *Registry) Disable(notification string) bool {
	return r.enableOrDisable(notification, false)
}

// SetDefault sets whether a notification is subscribed by default
// and reports whether it is in the supported list.
func (r *Registry) SetDefault(notification string, isDefault bool) bool {
	for i, supportedNotification := range r.supported {
		if notification == supportedNotification.Name {
			supportedNotification.Default = isDefault
			r.supported[i] = supportedNotification
			return true
		}
	}

	return false
}

// Supported returns the supported and enabled notifications.
func (r *Registry) Supported() []string {
	notifications := []string{}
	for _, supportedNotification := range r.supported {
		if supportedNotification.Enabled {
			notifications = append(notifications, supportedNotification.Name)
		}
	}

	return notifications
}

// Defaults returns the supported and enabled default notifications.
func (r *Registry) Defaults() []string {
	notifications := []string{}
	for _, supportedNotification := range r.supported {
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

func (r *Registry) enableOrDisable(notification string, enabled bool) bool {
	for i, supportedNotification := range r.supported {
		if notification == supportedNotification.Name {
			supportedNotification.Enabled = enabled
			r.supported[i] = supportedNotification
			return true
		}
	}

	return false
}
