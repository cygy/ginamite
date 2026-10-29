package notifications

import "context"

// respresents a notification that can be subscribed by a user.
type notificationType struct {
	Name    string
	Default bool
	Enabled bool
}

// Target : target of a notification to send a user.
type Target struct {
	Target string // a value defined into the file 'constants_targets.go'
	Value  string // an email address, a token, etc...
}

// GetNotificationTargetsByUserAndTypeFunc : func to get all the notification types subscribed by a user.
type GetNotificationTargetsByUserAndTypeFunc func(ctx context.Context, userID, notificationType string) ([]Target, error)
