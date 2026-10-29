package model

import (
	"context"
	"crypto/sha256"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UpdateNotifications : updates the notifications received by the user by email.
func (user *User) UpdateNotifications(ctx context.Context, notifications []string, db *mongo.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"notifications": notifications}}, db)
}

// UnsubscribeNotification : unsubscribe from a single notification.
func (user *User) UnsubscribeNotification(ctx context.Context, notificationType string, db *mongo.Database) error {
	return user.Update(ctx, bson.M{"$pull": bson.M{"notifications": notificationType}}, db)
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
func (user *User) UpdateCountOfNotifications(ctx context.Context, total, unread int64, db *mongo.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"notificationsStats.total": total, "notificationsStats.unread": unread}}, db)
}

// DecrementCountOfUnreadNotifications : decrements the count of unread notifications received by the user.
func (user *User) DecrementCountOfUnreadNotifications(ctx context.Context, db *mongo.Database) error {
	return user.Update(ctx, bson.M{"$inc": bson.M{"notificationsStats.unread": -1}}, db)
}
