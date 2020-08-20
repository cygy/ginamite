package model

import (
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UserNotificationTarget target of a notification to send a user.
type UserNotificationTarget struct {
	Type  string `bson:"type"`
	Value string `bson:"token"`
}

// GetUserNotificationTargetsByType : returns the targets of a given notification type of a user.
func GetUserNotificationTargetsByType(userID, notificationType string, db *mgo.Database) []UserNotificationTarget {
	targets := []UserNotificationTarget{}

	// Get the devices allowed to receive the notification.
	query := []bson.M{
		bson.M{"$match": bson.D{{
			Name:  "userId",
			Value: bson.ObjectIdHex(userID),
		}, {
			Name:  "device.notificationSettings.enabled",
			Value: true,
		}, {
			Name:  "device.notificationSettings.notifications",
			Value: bson.M{"$in": []string{notificationType}},
		}}},
		bson.M{"$replaceRoot": bson.M{"newRoot": "$device"}},
		bson.M{"$unwind": "$notificationSettings"},
	}

	pipe := db.C(AuthTokenCollection).Pipe(query)
	err := pipe.All(&targets)
	log.DatabaseError(err, UserCollection)

	// Get the email address if the user wants to receive this notification by email.
	user := &User{}
	user.GetByIDAndNotification(userID, notificationType, db)
	if user.IsSaved() && user.PrivateInfos.IsEmailValid {
		targets = append(targets, UserNotificationTarget{
			Type:  notifications.TargetEmail,
			Value: user.PrivateInfos.Email,
		})
	}

	return targets
}
