package account

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
)

// errNoNotifications is logged when a route of the notifications is called without storage of the notifications.
var errNoNotifications = errors.New("undefined storage of the notifications")

// GetNotifications returns the latest notifications received by the user.
func GetNotifications(c *gin.Context) {
	if StorageOf(c).Notifications == nil {
		c.Error(errNoNotifications)
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)

	notifications, err := StorageOf(c).Notifications.NotificationsReceivedByUser(c, userID, ctxLocale.Get(c))
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"notifications": notifications,
	})
}

// GetNotification returns a notification received by the user.
func GetNotification(c *gin.Context) {
	if StorageOf(c).Notifications == nil {
		c.Error(errNoNotifications)
		response.InternalServerError(c)
		return
	}

	errorMessageKey := "error.account.notification.get.message"

	notificationID, ok := checkNotificationID(c, errorMessageKey)
	if !ok {
		return
	}

	userID := ctxUser.GetID(c)

	notification, err := StorageOf(c).Notifications.NotificationReceivedByUser(c, userID, notificationID, ctxLocale.Get(c))
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.Ok(c, response.H{
		"notification": notification,
	})
}

// UpdateNotificationRead updates a notification object.
func UpdateNotificationRead(c *gin.Context) {
	errorMessageKey := "error.account.notification.read.update.message"

	notificationID, ok := checkNotificationID(c, errorMessageKey)
	if !ok {
		return
	}

	var err error
	if StorageOf(c).Notifications == nil {
		err = c.Error(errNoNotifications)
	} else {
		userID := ctxUser.GetID(c)
		err = StorageOf(c).Notifications.UpdateUserNotificationRead(c, userID, notificationID)
	}

	if err != nil {
		response.SendError(c, err)
		return
	}

	// The auth token is updated too with the latest properties of the user.
	ctxAuthToken.Refresh(c)

	locale := ctxLocale.Get(c)
	response.OkWithStatus(c, ctxkey.Localization.Get(c).Translate(locale)("status.account.notification.read.updated"))
}

// UpdateNotificationNotified updates a notification object.
func UpdateNotificationNotified(c *gin.Context) {
	errorMessageKey := "error.account.notification.notified.update.message"

	notificationID, ok := checkNotificationID(c, errorMessageKey)
	if !ok {
		return
	}

	var err error
	if StorageOf(c).Notifications == nil {
		err = c.Error(errNoNotifications)
	} else {
		userID := ctxUser.GetID(c)
		err = StorageOf(c).Notifications.UpdateUserNotificationNotified(c, userID, notificationID)
	}

	if err != nil {
		response.SendError(c, err)
		return
	}

	locale := ctxLocale.Get(c)
	response.OkWithStatus(c, ctxkey.Localization.Get(c).Translate(locale)("status.account.notification.notified.updated"))
}

// DeleteNotification deletes a notification object.
func DeleteNotification(c *gin.Context) {
	errorMessageKey := "error.account.notification.delete.message"

	notificationID, ok := checkNotificationID(c, errorMessageKey)
	if !ok {
		return
	}

	var err error
	if StorageOf(c).Notifications == nil {
		err = c.Error(errNoNotifications)
	} else {
		userID := ctxUser.GetID(c)
		err = StorageOf(c).Notifications.DeleteUserNotification(c, userID, notificationID)
	}

	if err != nil {
		response.SendError(c, err)
		return
	}

	// The auth token is updated too with the latest properties of the user.
	ctxAuthToken.Refresh(c)

	locale := ctxLocale.Get(c)
	response.OkWithStatus(c, ctxkey.Localization.Get(c).Translate(locale)("status.account.notification.deleted"))
}

// GetNotificationsSettings returns all the notifications received or not by the user's devices.
func GetNotificationsSettings(c *gin.Context) {
	if StorageOf(c).Notifications == nil {
		c.Error(errNoNotifications)
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)
	userNotifications, err := StorageOf(c).Notifications.NotificationsSubscribedByUser(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"notifications": userNotifications,
	})
}

// SaveNotificationsSettings saves all the notifications received by the user's devices.
func SaveNotificationsSettings(c *gin.Context) {
	if StorageOf(c).Notifications == nil {
		c.Error(errNoNotifications)
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Notifications []string `json:"notifications"`
	}
	if !request.DecodeBody(c, &jsonBody) {
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	notificationsToSave := jsonBody.Notifications
	for _, notification := range notificationsToSave {
		if !ctxkey.Notifications.Get(c).IsSupported(notification) {
			message := t("error.notifications.update.unable.message")
			reason := t("error.notifications.update.invalid.reason", localization.H{"Value": notification})
			recovery := t("error.notifications.update.invalid.recovery", localization.H{"Value": strings.Join(ctxkey.Notifications.Get(c).Supported(), ",")})

			response.InvalidParameterValue(c, "notifications", message, reason, recovery)
			return
		}
	}

	userID := ctxUser.GetID(c)
	if err := StorageOf(c).Notifications.SubscribeNotificationsToUser(c, userID, notificationsToSave); err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.notifications.updated"))
}

// UnsubscribeNotificationByKey unsubscribes from an encrypted key.
func UnsubscribeNotificationByKey(c *gin.Context) {
	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// The same error is always sent to avoid giving some clues.
	sendError := func() {
		message := t("error.notification.unsubscribe.message")
		reason := t("error.notification.unsubscribe.reason")
		recovery := t("error.notification.unsubscribe.recovery")
		response.InvalidRequestParameterWithDetails(c, message, reason, recovery)
	}

	if StorageOf(c).Notifications == nil {
		c.Error(errNoNotifications)
		sendError()
		return
	}

	// key is mandatory.
	key := c.Param("key")
	if len(key) == 0 {
		sendError()
		return
	}

	// Check the composition of the encoded key.
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		sendError()
		return
	}

	// Check the userID.
	userID := parts[0]

	// Verify that the notification is supported.
	notification := parts[1]
	if !ctxkey.Notifications.Get(c).IsSupported(notification) {
		sendError()
		return
	}

	if err := StorageOf(c).Notifications.UnsubscribeNotificationFromUser(c, userID, notification, key); err != nil {
		sendError()
		return
	}

	response.OkWithStatus(c, t("status.notification.unsubscribed"))
}
