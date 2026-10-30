package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// GetUserNotifications : returns the latest notifications received by a user.
func GetUserNotifications(ctx context.Context, userID string, db store.Database) ([]Notification, error) {
	notifications := []Notification{}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return notifications, err
	}

	err = document.GetDocumentsBySelectorAndSort(ctx, &notifications, NotificationCollection, bson.M{"userId": objectID}, []string{"userId", "-createdAt"}, db)
	return notifications, err
}

// GetCountOfUserNotifications : returns the count of notifications received by a user.
func GetCountOfUserNotifications(ctx context.Context, userID string, db store.Database) (int64, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	return document.CountBySelector(ctx, NotificationCollection, bson.M{"userId": objectID}, db)
}

// GetCountOfUserUnreadNotifications : returns the count of unread notifications received by a user.
func GetCountOfUserUnreadNotifications(ctx context.Context, userID string, db store.Database) (int64, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}

	return document.CountBySelector(ctx, NotificationCollection, bson.D{
		{Key: "userId", Value: objectID},
		{Key: "read", Value: false},
	}, db)
}

// GetOldestUserNotifications : returns the IDs of the oldest notifications received by a user.
func GetOldestUserNotifications(ctx context.Context, userID string, offset int64, db store.Database) ([]bson.ObjectID, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return []bson.ObjectID{}, err
	}

	notifications := []Notification{}
	err = document.GetDocumentsBySelectorAndSortWithOffsetAndLimit(ctx, &notifications, NotificationCollection, bson.M{"userId": objectID}, []string{"userId", "-createdAt"}, offset, -1, db)

	notificationIDs := make([]bson.ObjectID, len(notifications))
	for i, notification := range notifications {
		notificationIDs[i] = notification.ID
	}

	return notificationIDs, err
}

// DeleteUserNotifications : deletes some notifications received by a user.
func DeleteUserNotifications(ctx context.Context, notificationsIDs []bson.ObjectID, db store.Database) (int64, error) {
	IDs := make([]bson.M, len(notificationsIDs))
	for i, notificationsID := range notificationsIDs {
		IDs[i] = bson.M{"_id": notificationsID}
	}

	return document.DeleteDocumentsBySelector(ctx, NotificationCollection, bson.M{"$or": IDs}, db)
}
