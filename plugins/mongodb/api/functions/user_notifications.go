package functions

import (
	"context"
	"crypto/subtle"
	"errors"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// Notifications stores the notifications of the users in a database. It is an account.NotificationsStorage. Create it
// with NewNotifications.
type Notifications struct {
	database      func() store.Database
	defaultLocale func() string
}

// NewNotifications returns the storage of the notifications working on the database returned by the function, which is
// called each time the storage needs the database. The texts of the notifications fall back to the default locale,
// returned by the second function.
func NewNotifications(database func() store.Database, defaultLocale func() string) *Notifications {
	return &Notifications{database: database, defaultLocale: defaultLocale}
}

// NotificationsReceivedByUser returns the latest notifications received by the user.
func (s *Notifications) NotificationsReceivedByUser(ctx context.Context, userID, locale string) (any, error) {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return nil, err
	}

	notifications, err := model.GetUserNotifications(ctx, userID, s.database())
	if err != nil {
		return nil, err
	}

	defaultLocale := s.defaultLocale()
	summarizedNotifications := make([]model.OwnSummarizedNotification, len(notifications))
	for i, notification := range notifications {
		summarizedNotifications[i] = model.NewOwnSummarizedNotification(notification, locale, defaultLocale)
	}

	return summarizedNotifications, nil
}

// NotificationReceivedByUser returns a notification received by the user.
func (s *Notifications) NotificationReceivedByUser(ctx context.Context, userID, notificationID, locale string) (any, error) {
	notification, err := s.userNotification(ctx, userID, notificationID)
	if err != nil {
		return nil, err
	}

	return model.NewOwnNotification(*notification, locale, s.defaultLocale()), nil
}

// UpdateUserNotificationRead updates the read flag of a notification received by the user.
func (s *Notifications) UpdateUserNotificationRead(ctx context.Context, userID, notificationID string) error {
	notification, err := s.userNotification(ctx, userID, notificationID)
	if err != nil {
		return err
	}

	if notification.Read {
		return nil
	}

	db := s.database()
	if err := notification.UpdateRead(ctx, db); err != nil {
		return err
	}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	user := model.User{}
	user.ID = objectID
	if err := user.DecrementCountOfUnreadNotifications(ctx, db); err != nil {
		log.WithField("error", err).Error("unable to decrement count of unread notifications")
	}

	return nil
}

// UpdateUserNotificationNotified updates the notified flag of a notification received by the user.
func (s *Notifications) UpdateUserNotificationNotified(ctx context.Context, userID, notificationID string) error {
	notification, err := s.userNotification(ctx, userID, notificationID)
	if err != nil {
		return err
	}

	if notification.Notified {
		return nil
	}

	return notification.UpdateNotified(ctx, s.database())
}

// DeleteUserNotification deletes a notification received by the user.
func (s *Notifications) DeleteUserNotification(ctx context.Context, userID, notificationID string) error {
	notification, err := s.userNotification(ctx, userID, notificationID)
	if err != nil {
		return err
	}

	db := s.database()
	if err := notification.Delete(ctx, db); err != nil {
		return err
	}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Update the count of user notifications, total and unread.
	countOfNotifications, err := model.GetCountOfUserNotifications(ctx, userID, db)
	if err != nil {
		return err
	}
	countOfUnreadNotifications, err := model.GetCountOfUserUnreadNotifications(ctx, userID, db)
	if err != nil {
		return err
	}

	user := model.User{}
	user.ID = objectID
	if err := user.UpdateCountOfNotifications(ctx, countOfNotifications, countOfUnreadNotifications, db); err != nil {
		log.WithField("error", err).Error("unable to update count of notifications")
	}

	return nil
}

// SubscribeNotificationsToUser sets the notifications received by the user by email.
func (s *Notifications) SubscribeNotificationsToUser(ctx context.Context, userID string, notifications []string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateNotifications(ctx, notifications, s.database())
}

// UnsubscribeNotificationFromUser unsubscribes the user from a notification, if the unsubscribe key matches.
func (s *Notifications) UnsubscribeNotificationFromUser(ctx context.Context, userID, notification, unsubscribeKey string) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	// Check the encoded key.
	db := s.database()
	user := &model.User{}
	if err := user.GetByID(ctx, userID, db); err != nil {
		return err
	}

	// Validate the encoded key.
	verifiedUnsubscribeKey := notifications.GenerateUnsubscribeKey(userID, user.PrivateKey, notification)
	if subtle.ConstantTimeCompare([]byte(verifiedUnsubscribeKey), []byte(unsubscribeKey)) != 1 {
		return errors.New("wrong unsubscribe key")
	}

	return user.UnsubscribeNotification(ctx, notification, db)
}

// NotificationsSubscribedByUser returns the notifications subscribed by the user, none if the user does not exist.
func (s *Notifications) NotificationsSubscribedByUser(ctx context.Context, userID string) ([]string, error) {
	user := &model.User{}
	if err := user.GetByID(ctx, userID, s.database()); err != nil {
		if model.IsNotFound(err) {
			return []string{}, nil
		}

		return nil, err
	}

	return user.Notifications, nil
}

// userNotification returns the notification of the user, or a "not found" error if it belongs to another user.
func (s *Notifications) userNotification(ctx context.Context, userID, notificationID string) (*model.Notification, error) {
	// Check the userID and the notificationID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return nil, err
	}
	if _, err := bson.ObjectIDFromHex(notificationID); err != nil {
		return nil, err
	}

	notification := &model.Notification{}
	if err := notification.GetByID(ctx, notificationID, s.database()); err != nil {
		return nil, err
	}

	if notification.UserID.Hex() != userID {
		return nil, commonErrors.NotFound()
	}

	return notification, nil
}
