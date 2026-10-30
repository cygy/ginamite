package model

import (
	"context"
	"errors"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/common/notifications"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UserNotificationTarget target of a notification to send a user.
type UserNotificationTarget struct {
	Type  string `bson:"type"`
	Value string `bson:"token"`
}

// GetUserNotificationTargetsByType : returns the targets of a given notification type of a user.
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
		log.DatabaseError(err, UserCollection)
		return targets, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.WithField("error", err).Error("unable to close the cursor")
		}
	}()

	err = cursor.All(ctx, &targets)
	if err != nil {
		log.DatabaseError(err, UserCollection)
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
