package account

import "context"

// NotificationsStorage is the storage of the notifications of the users, provided by the application, such as the one
// of the storage plugin. The context is the one of the request.
type NotificationsStorage interface {
	// NotificationsReceivedByUser returns the latest notifications received by the user, with the texts in the locale.
	NotificationsReceivedByUser(ctx context.Context, userID, locale string) (any, error)
	// NotificationReceivedByUser returns a notification received by the user, with the texts in the locale.
	NotificationReceivedByUser(ctx context.Context, userID, notificationID, locale string) (any, error)
	// UpdateUserNotificationRead marks a notification received by the user as read.
	UpdateUserNotificationRead(ctx context.Context, userID, notificationID string) error
	// UpdateUserNotificationNotified marks a notification received by the user as notified.
	UpdateUserNotificationNotified(ctx context.Context, userID, notificationID string) error
	// DeleteUserNotification deletes a notification received by the user.
	DeleteUserNotification(ctx context.Context, userID, notificationID string) error
	// NotificationsSubscribedByUser returns the notifications subscribed by the user.
	NotificationsSubscribedByUser(ctx context.Context, userID string) ([]string, error)
	// SubscribeNotificationsToUser replaces the notifications subscribed by the user.
	SubscribeNotificationsToUser(ctx context.Context, userID string, notifications []string) error
	// UnsubscribeNotificationFromUser unsubscribes the user from a notification, if the unsubscribe key is valid.
	UnsubscribeNotificationFromUser(ctx context.Context, userID, notification, unsubscribeKey string) error
}
