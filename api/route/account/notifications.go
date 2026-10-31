package account

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/common/ctxnotifications"

	"github.com/cygy/ginamite/common/ctxlocalization"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
)

// GetNotifications returns the latest notifications received by the user.
func GetNotifications(c *gin.Context) {
	if FunctionsOf(c).NotificationsReceivedByUser == nil {
		c.Error(errors.New("undefined function 'NotificationsReceivedByUser'"))
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)

	notifications, err := FunctionsOf(c).NotificationsReceivedByUser(c, userID)
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
	if FunctionsOf(c).NotificationReceivedByUser == nil {
		c.Error(errors.New("undefined function 'NotificationReceivedByUser'"))
		response.InternalServerError(c)
		return
	}

	errorMessageKey := "error.account.notification.get.message"

	notificationID, ok := checkNotificationID(c, errorMessageKey)
	if !ok {
		return
	}

	userID := ctxUser.GetID(c)

	notification, err := FunctionsOf(c).NotificationReceivedByUser(c, userID, notificationID)
	if err != nil {
		response.InternalServerError(c)
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
	if FunctionsOf(c).UpdateUserNotificationRead == nil {
		err = c.Error(errors.New("undefined function 'UpdateUserNotificationRead'"))
	} else {
		userID := ctxUser.GetID(c)
		err = FunctionsOf(c).UpdateUserNotificationRead(c, userID, notificationID)
	}

	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The auth token is updated too with the latest properties of the user.
	ctxAuthToken.Refresh(c)

	locale := ctxLocale.Get(c)
	response.OkWithStatus(c, ctxlocalization.Get(c).Translate(locale)("status.account.notification.read.updated"))
}

// UpdateNotificationNotified updates a notification object.
func UpdateNotificationNotified(c *gin.Context) {
	errorMessageKey := "error.account.notification.notified.update.message"

	notificationID, ok := checkNotificationID(c, errorMessageKey)
	if !ok {
		return
	}

	var err error
	if FunctionsOf(c).UpdateUserNotificationNotified == nil {
		err = c.Error(errors.New("undefined function 'UpdateUserNotificationNotified'"))
	} else {
		userID := ctxUser.GetID(c)
		err = FunctionsOf(c).UpdateUserNotificationNotified(c, userID, notificationID)
	}

	if err != nil {
		response.InternalServerError(c)
		return
	}

	locale := ctxLocale.Get(c)
	response.OkWithStatus(c, ctxlocalization.Get(c).Translate(locale)("status.account.notification.notified.updated"))
}

// DeleteNotification deletes a notification object.
func DeleteNotification(c *gin.Context) {
	errorMessageKey := "error.account.notification.delete.message"

	notificationID, ok := checkNotificationID(c, errorMessageKey)
	if !ok {
		return
	}

	var err error
	if FunctionsOf(c).DeleteUserNotification == nil {
		err = c.Error(errors.New("undefined function 'DeleteUserNotification'"))
	} else {
		userID := ctxUser.GetID(c)
		err = FunctionsOf(c).DeleteUserNotification(c, userID, notificationID)
	}

	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The auth token is updated too with the latest properties of the user.
	ctxAuthToken.Refresh(c)

	locale := ctxLocale.Get(c)
	response.OkWithStatus(c, ctxlocalization.Get(c).Translate(locale)("status.account.notification.deleted"))
}

// GetNotificationsSettings returns all the notifications received or not by the user's devices.
func GetNotificationsSettings(c *gin.Context) {
	if FunctionsOf(c).NotificationsSubscribedByUser == nil {
		c.Error(errors.New("undefined function 'NotificationsSubscribedByUser'"))
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)
	userNotifications := FunctionsOf(c).NotificationsSubscribedByUser(c, userID)
	response.Ok(c, response.H{
		"notifications": userNotifications,
	})
}

// SaveNotificationsSettings saves all the notifications received by the user's devices.
func SaveNotificationsSettings(c *gin.Context) {
	if FunctionsOf(c).SubscribeNotificationsToUser == nil {
		c.Error(errors.New("undefined function 'SubscribeNotificationsToUser'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Notifications []string `json:"notifications"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxlocalization.Get(c).Translate(locale)

	notificationsToSave := jsonBody.Notifications
	for _, notification := range notificationsToSave {
		if !ctxnotifications.Get(c).IsSupported(notification) {
			message := t("error.notifications.update.unable.message")
			reason := t("error.notifications.update.invalid.reason", localization.H{"Value": notification})
			recovery := t("error.notifications.update.invalid.recovery", localization.H{"Value": strings.Join(ctxnotifications.Get(c).Supported(), ",")})

			response.InvalidParameterValue(c, "notifications", message, reason, recovery)
			return
		}
	}

	userID := ctxUser.GetID(c)
	if err := FunctionsOf(c).SubscribeNotificationsToUser(c, userID, notificationsToSave); err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.notifications.updated"))
}

// UnsubscribeNotificationByKey unsubscribes from an encrypted key.
func UnsubscribeNotificationByKey(c *gin.Context) {
	locale := ctxLocale.Get(c)
	t := ctxlocalization.Get(c).Translate(locale)

	// The same error is always sent to avoid giving some clues.
	sendError := func() {
		message := t("error.notification.unsubscribe.message")
		reason := t("error.notification.unsubscribe.reason")
		recovery := t("error.notification.unsubscribe.recovery")
		response.InvalidRequestParameterWithDetails(c, message, reason, recovery)
	}

	if FunctionsOf(c).UnsubscribeNotificationFromUser == nil {
		c.Error(errors.New("undefined function 'UnsubscribeNotificationFromUser'"))
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
	if !ctxnotifications.Get(c).IsSupported(notification) {
		sendError()
		return
	}

	if err := FunctionsOf(c).UnsubscribeNotificationFromUser(c, userID, notification, key); err != nil {
		sendError()
		return
	}

	response.OkWithStatus(c, t("status.notification.unsubscribed"))
}
