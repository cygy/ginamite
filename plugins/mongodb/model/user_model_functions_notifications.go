package model

import (
	"crypto/sha256"
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// UpdateNotifications : updates the notifications received by the user by email.
func (user *User) UpdateNotifications(notifications []string, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"notifications": notifications}}, db)
}

// UnsubscribeNotification : unsubscribe from a single notification.
func (user *User) UnsubscribeNotification(notificationType string, db *mgo.Database) error {
	return user.Update(bson.M{"$pull": bson.M{"notifications": notificationType}}, db)
}

// GenerateUnsubscribeKey : generates an unsubscribe key. This key is using to unsubscribe a notification by email without authorization.
func (user *User) GenerateUnsubscribeKey(notificationType string) string {
	userID := user.ID.Hex()
	privateKey := user.PrivateKey

	sha256 := sha256.New()
	sha256.Write([]byte(userID + notificationType + privateKey))
	encodedKey := sha256.Sum(nil)

	return fmt.Sprintf("%s.%s.%s", userID, notificationType, fmt.Sprintf("%x", encodedKey))
}

// UpdateCountOfNotifications : updates the count of notifications received by the user.
func (user *User) UpdateCountOfNotifications(total, unread int, db *mgo.Database) error {
	return user.Update(bson.M{"$set": bson.M{"notificationsStats.total": total, "notificationsStats.unread": unread}}, db)
}

// DecrementCountOfUnreadNotifications : decrements the count of unread notifications received by the user.
func (user *User) DecrementCountOfUnreadNotifications(db *mgo.Database) error {
	return user.Update(bson.M{"$inc": bson.M{"notificationsStats.unread": -1}}, db)
}
