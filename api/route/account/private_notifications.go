package account

import (
	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

func checkNotificationID(c *gin.Context, messageKey string) (string, bool) {
	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	notificationID := c.Param("id")
	if len(notificationID) == 0 {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.account.notification.id.not_found.reason"),
			t("error.account.notification.id.not_found.recovery"))
		return "", false
	}

	if !bson.IsObjectIdHex(notificationID) {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.account.notification.id.invalid.reason"),
			t("error.account.notification.id.invalid.recovery"))
		return "", false
	}

	return notificationID, true
}
