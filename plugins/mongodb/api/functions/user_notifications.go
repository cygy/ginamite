package functions

import (
	"errors"

	apiContext "github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// NotificationsReceivedByUser : returns the latest notifications received by the user.
func NotificationsReceivedByUser(c *gin.Context, userID string) (interface{}, error) {
	// Check the userID.
	if !bson.IsObjectIdHex(userID) {
		return nil, errors.New("wrong object ID")
	}

	mongoSession := session.Get(c)

	notifications, err := model.GetUserNotifications(userID, mongoSession)
	if err != nil {
		return nil, err
	}

	locale := apiContext.GetLocale(c)

	summarizedNotifications := make([]model.OwnSummarizedNotification, len(notifications))
	for i, notification := range notifications {
		summarizedNotifications[i] = model.NewOwnSummarizedNotification(notification, locale, config.Main.DefaultLocale)
	}

	return summarizedNotifications, nil
}

// NotificationReceivedByUser : returns a notification received by the user.
func NotificationReceivedByUser(c *gin.Context, userID, notificationID string) (interface{}, error) {
	notification, err := checkUserIDAndNotificationID(c, userID, notificationID)
	if err != nil {
		return nil, err
	}

	locale := apiContext.GetLocale(c)

	return model.NewOwnNotification(*notification, locale, config.Main.DefaultLocale), nil
}

// UpdateUserNotificationRead : updates the read flag of a notification received by the user.
func UpdateUserNotificationRead(c *gin.Context, userID, notificationID string) error {
	notification, err := checkUserIDAndNotificationID(c, userID, notificationID)
	if err != nil {
		return err
	}

	if notification.Read {
		return nil
	}

	mongoSession := session.Get(c)

	if err := notification.UpdateRead(mongoSession); err != nil {
		return err
	}

	user := model.User{}
	user.ID = bson.ObjectIdHex(userID)
	user.DecrementCountOfUnreadNotifications(mongoSession)

	return nil
}

// UpdateUserNotificationNotified : updates the notified flag of a notification received by the user.
func UpdateUserNotificationNotified(c *gin.Context, userID, notificationID string) error {
	notification, err := checkUserIDAndNotificationID(c, userID, notificationID)
	if err != nil {
		return err
	}

	if notification.Notified {
		return nil
	}

	mongoSession := session.Get(c)

	return notification.UpdateNotified(mongoSession)
}

// DeleteUserNotification : deletes a notification received by the user.
func DeleteUserNotification(c *gin.Context, userID, notificationID string) error {
	notification, err := checkUserIDAndNotificationID(c, userID, notificationID)
	if err != nil {
		return err
	}

	mongoSession := session.Get(c)

	if err := notification.Delete(mongoSession); err != nil {
		return err
	}

	// Update the count of user notifications, total and unread.
	countOfNotifications, err := model.GetCountOfUserNotifications(userID, mongoSession)
	if err != nil {
		return err
	}

	countOfUnreadNotifications, err := model.GetCountOfUserUnreadNotifications(userID, mongoSession)
	if err != nil {
		return err
	}

	user := model.User{}
	user.ID = bson.ObjectIdHex(userID)
	user.UpdateCountOfNotifications(countOfNotifications, countOfUnreadNotifications, mongoSession)

	return nil
}

// SubscribeNotificationsToUser : the user subscribes to these notifications.
func SubscribeNotificationsToUser(c *gin.Context, userID string, notifications []string) error {
	mongoSession := session.Get(c)
	user := context.GetUser(c)
	return user.UpdateNotifications(notifications, mongoSession)
}

// UnsubscribeNotificationFromUser : the user has unsubscribed from a notification.
func UnsubscribeNotificationFromUser(c *gin.Context, userID, notification, unsubscribeKey string) error {
	// Check the userID.
	if !bson.IsObjectIdHex(userID) {
		return errors.New("wrong object ID")
	}

	// Check the encoded key.
	mongoSession := session.Get(c)
	user := &model.User{}
	if err := user.GetByID(userID, mongoSession); err != nil {
		return err
	}

	// Validate the encoded key.
	verifiedUnsubscribeKey := notifications.GenerateUnsubscribeKey(userID, user.PrivateKey, notification)
	if verifiedUnsubscribeKey != unsubscribeKey {
		return errors.New("wrong unsubscribe key")
	}

	if err := user.UnsubscribeNotification(notification, mongoSession); err != nil {
		return err
	}

	return nil
}

// NotificationsSubscribedByUser : returns the notifications subscribed by the user.
func NotificationsSubscribedByUser(c *gin.Context, userID string) []string {
	user := context.GetFullUser(c)
	return user.Notifications
}
