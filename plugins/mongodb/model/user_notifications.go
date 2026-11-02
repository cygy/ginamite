package model

import (
	"context"
	"errors"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UpdateNotifications updates the notifications received by the user by email.
func (user *User) UpdateNotifications(ctx context.Context, notifications []string, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"notifications": notifications}}, db)
}

// UnsubscribeNotification unsubscribes from a single notification.
func (user *User) UnsubscribeNotification(ctx context.Context, notificationType string, db store.Database) error {
	return user.Update(ctx, bson.M{"$pull": bson.M{"notifications": notificationType}}, db)
}

// GenerateUnsubscribeKey returns the key to unsubscribe the user from a notification by email, without being signed in.
// The key is derived from the private key of the user.
func (user *User) GenerateUnsubscribeKey(notificationType string) string {
	return notifications.GenerateUnsubscribeKey(user.ID.Hex(), user.PrivateKey, notificationType)
}

// UpdateCountOfNotifications updates the count of notifications received by the user.
func (user *User) UpdateCountOfNotifications(ctx context.Context, total, unread int64, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"notificationsStats.total": total, "notificationsStats.unread": unread}}, db)
}

// DecrementCountOfUnreadNotifications decrements the count of unread notifications received by the user.
func (user *User) DecrementCountOfUnreadNotifications(ctx context.Context, db store.Database) error {
	return user.Update(ctx, bson.M{"$inc": bson.M{"notificationsStats.unread": -1}}, db)
}

// UserNotificationTarget is a target of a notification sent to a user: a device token or an email address.
type UserNotificationTarget struct {
	Type  string `bson:"type"`
	Value string `bson:"token"`
}

// GetUserNotificationTargetsByType returns the targets of a given notification type of a user.
func GetUserNotificationTargetsByType(ctx context.Context, userID, notificationType string, db store.Database) ([]UserNotificationTarget, error) {
	targets := []UserNotificationTarget{}

	var objectID bson.ObjectID
	var err error
	if objectID, err = bson.ObjectIDFromHex(userID); err != nil {
		return targets, err
	}

	// Get the devices allowed to receive the notification.
	pipeline := []bson.M{
		{"$match": bson.D{{
			Key:   "userId",
			Value: objectID,
		}, {
			Key:   "device.notificationSettings.enabled",
			Value: true,
		}, {
			Key:   "device.notificationSettings.notifications",
			Value: bson.M{"$in": []string{notificationType}},
		}}},
		{"$replaceRoot": bson.M{"newRoot": "$device"}},
		{"$unwind": "$notificationSettings"},
	}

	cursor, err := db.Collection(AuthTokenCollection).Aggregate(ctx, pipeline)
	if err != nil {
		log.DatabaseError(err, AuthTokenCollection)
		return targets, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.WithField("error", err).Error("unable to close the cursor")
		}
	}()

	err = cursor.All(ctx, &targets)
	if err != nil {
		log.DatabaseError(err, AuthTokenCollection)
		return targets, err
	}

	// Get the email address if the user wants to receive this notification by email.
	user := &User{}
	if err := user.GetByIDAndNotification(ctx, userID, notificationType, db); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return targets, err
	}
	if user.IsSaved() && user.PrivateInfos.IsEmailValid {
		targets = append(targets, UserNotificationTarget{
			Type:  notifications.TargetEmail,
			Value: user.PrivateInfos.Email,
		})
	}

	return targets, nil
}
