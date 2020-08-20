package model

import (
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// GetUserNotifications : returns the latest notifications received by a user.
func GetUserNotifications(userID string, db *mgo.Database) ([]Notification, error) {
	notifications := []Notification{}
	err := document.GetDocumentsBySelectorAndSort(&notifications, NotificationCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, []string{"userId", "-createdAt"}, db)

	return notifications, err
}

// GetCountOfUserNotifications : returns the count of notifications received by a user.
func GetCountOfUserNotifications(userID string, db *mgo.Database) (int, error) {
	return document.CountBySelector(NotificationCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, db)
}

// GetCountOfUserUnreadNotifications : returns the count of unread notifications received by a user.
func GetCountOfUserUnreadNotifications(userID string, db *mgo.Database) (int, error) {
	return document.CountBySelector(NotificationCollection, bson.D{{Name: "userId", Value: bson.ObjectIdHex(userID)}, {Name: "read", Value: false}}, db)
}

// GetOldestUserNotifications : returns the IDs of the oldest notifications received by a user.
func GetOldestUserNotifications(userID string, offset int, db *mgo.Database) ([]bson.ObjectId, error) {
	notifications := []Notification{}
	err := document.GetDocumentsBySelectorAndSortWithOffsetAndLimit(&notifications, NotificationCollection, bson.M{"userId": bson.ObjectIdHex(userID)}, []string{"userId", "-createdAt"}, offset, -1, db)

	notificationIDs := make([]bson.ObjectId, len(notifications))
	for i, notification := range notifications {
		notificationIDs[i] = notification.ID
	}

	return notificationIDs, err
}

// DeleteUserNotifications : deletes some notifications received by a user.
func DeleteUserNotifications(notificationsIDs []bson.ObjectId, db *mgo.Database) (int, error) {
	IDs := make([]bson.M, len(notificationsIDs))
	for i, notificationsID := range notificationsIDs {
		IDs[i] = bson.M{"_id": notificationsID}
	}

	return document.DeleteDocumentsBySelector(NotificationCollection, bson.M{"$or": IDs}, db)
}
