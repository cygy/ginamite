package account

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/authentication"
)

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

// TokensStorage is the storage of the auth tokens, provided by the application, such as the one of the storage plugin.
// The context is the one of the request.
type TokensStorage interface {
	// SaveAuthenticationToken saves an auth token and returns its ID, and whether the IP address is already known.
	SaveAuthenticationToken(ctx context.Context, details authentication.Details) (tokenID string, knownIPAddress bool, err error)
	// GetTokenExpirationDate returns the expiration date of an auth token.
	GetTokenExpirationDate(ctx context.Context, tokenID string) (time.Time, error)
	// DeleteTokenByID deletes an auth token.
	DeleteTokenByID(ctx context.Context, tokenID string) error
	// GetTokenDetailsByID returns the details of an auth token.
	GetTokenDetailsByID(ctx context.Context, tokenID string) (authentication.Token, error)
	// GetOwnedTokens returns the auth tokens of the user, to display them to the user.
	GetOwnedTokens(ctx context.Context, userID string) (any, error)
	// GetOwnedTokenByID returns an auth token, to display it to its owner, and the ID of its owner.
	GetOwnedTokenByID(ctx context.Context, tokenID string) (token any, ownerID string, err error)
	// UpdateTokenByID updates the name and the notifications of an auth token.
	UpdateTokenByID(ctx context.Context, tokenID, name string, enableNotifications bool, notifications []string) error
}
