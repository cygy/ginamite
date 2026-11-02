package functions

import (
	"crypto/subtle"
	"errors"

	apiContextLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/common/ctxkey"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"

	"github.com/gin-gonic/gin"
)

// NotificationsReceivedByUser returns the latest notifications received by the user.
func NotificationsReceivedByUser(c *gin.Context, userID string) (any, error) {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return nil, err
	}

	mongoSession := session.Get(c)

	notifications, err := model.GetUserNotifications(c, userID, mongoSession)
	if err != nil {
		return nil, err
	}

	locale := apiContextLocale.Get(c)

	summarizedNotifications := make([]model.OwnSummarizedNotification, len(notifications))
	for i, notification := range notifications {
		summarizedNotifications[i] = model.NewOwnSummarizedNotification(notification, locale, ctxkey.Config.Get(c).DefaultLocale)
	}

	return summarizedNotifications, nil
}

// NotificationReceivedByUser returns a notification received by the user.
func NotificationReceivedByUser(c *gin.Context, userID, notificationID string) (any, error) {
	notification, err := checkUserIDAndNotificationID(c, userID, notificationID)
	if err != nil {
		return nil, err
	}

	locale := apiContextLocale.Get(c)

	return model.NewOwnNotification(*notification, locale, ctxkey.Config.Get(c).DefaultLocale), nil
}

// UpdateUserNotificationRead updates the read flag of a notification received by the user.
func UpdateUserNotificationRead(c *gin.Context, userID, notificationID string) error {
	notification, err := checkUserIDAndNotificationID(c, userID, notificationID)
	if err != nil {
		return err
	}

	if notification.Read {
		return nil
	}

	mongoSession := session.Get(c)

	if err := notification.UpdateRead(c, mongoSession); err != nil {
		return err
	}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	user := model.User{}
	user.ID = objectID
	if err := user.DecrementCountOfUnreadNotifications(c, mongoSession); err != nil {
		log.WithField("error", err).Error("unable to decrement count of unread notifications")
	}

	return nil
}

// UpdateUserNotificationNotified updates the notified flag of a notification received by the user.
func UpdateUserNotificationNotified(c *gin.Context, userID, notificationID string) error {
	notification, err := checkUserIDAndNotificationID(c, userID, notificationID)
	if err != nil {
		return err
	}

	if notification.Notified {
		return nil
	}

	mongoSession := session.Get(c)

	return notification.UpdateNotified(c, mongoSession)
}

// DeleteUserNotification deletes a notification received by the user.
func DeleteUserNotification(c *gin.Context, userID, notificationID string) error {
	notification, err := checkUserIDAndNotificationID(c, userID, notificationID)
	if err != nil {
		return err
	}

	mongoSession := session.Get(c)

	if err := notification.Delete(c, mongoSession); err != nil {
		return err
	}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Update the count of user notifications, total and unread.
	countOfNotifications, err := model.GetCountOfUserNotifications(c, userID, mongoSession)
	if err != nil {
		return err
	}
	countOfUnreadNotifications, err := model.GetCountOfUserUnreadNotifications(c, userID, mongoSession)
	if err != nil {
		return err
	}

	user := model.User{}
	user.ID = objectID
	if err := user.UpdateCountOfNotifications(c, countOfNotifications, countOfUnreadNotifications, mongoSession); err != nil {
		log.WithField("error", err).Error("unable to update count of notifications")
	}

	return nil
}

// SubscribeNotificationsToUser sets the notifications received by the authenticated user by email.
func SubscribeNotificationsToUser(c *gin.Context, userID string, notifications []string) error {
	mongoSession := session.Get(c)
	user := context.GetUser(c)
	return user.UpdateNotifications(c, notifications, mongoSession)
}

// UnsubscribeNotificationFromUser unsubscribes the user from a notification, if the unsubscribe key matches.
func UnsubscribeNotificationFromUser(c *gin.Context, userID, notification, unsubscribeKey string) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	// Check the encoded key.
	mongoSession := session.Get(c)
	user := &model.User{}
	if err := user.GetByID(c, userID, mongoSession); err != nil {
		return err
	}

	// Validate the encoded key.
	verifiedUnsubscribeKey := notifications.GenerateUnsubscribeKey(userID, user.PrivateKey, notification)
	if subtle.ConstantTimeCompare([]byte(verifiedUnsubscribeKey), []byte(unsubscribeKey)) != 1 {
		return errors.New("wrong unsubscribe key")
	}

	if err := user.UnsubscribeNotification(c, notification, mongoSession); err != nil {
		return err
	}

	return nil
}

// NotificationsSubscribedByUser returns the notifications subscribed by the user.
func NotificationsSubscribedByUser(c *gin.Context, userID string) []string {
	user := context.GetFullUser(c)
	if user == nil {
		return []string{}
	}

	return user.Notifications
}

func checkUserIDAndNotificationID(c *gin.Context, userID, notificationID string) (*model.Notification, error) {
	// Check the userID and the notificationID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return nil, err
	}
	if _, err := bson.ObjectIDFromHex(notificationID); err != nil {
		return nil, err
	}

	mongoSession := session.Get(c)

	notification := &model.Notification{}
	if err := notification.GetByID(c, notificationID, mongoSession); err != nil {
		return nil, err
	}

	if notification.UserID.Hex() != userID {
		return nil, commonErrors.NotFound()
	}

	return notification, nil
}
