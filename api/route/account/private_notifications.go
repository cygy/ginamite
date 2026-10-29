package account

import (
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/gin-gonic/gin"
)

func checkNotificationID(c *gin.Context, messageKey string) (string, bool) {
	locale := ctxLocale.GetLocale(c)
	t := localization.Translate(locale)

	notificationID := c.Param("id")
	if len(notificationID) == 0 {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.account.notification.id.not_found.reason"),
			t("error.account.notification.id.not_found.recovery"))
		return "", false
	}

	if _, err := bson.ObjectIDFromHex(notificationID); err != nil {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.account.notification.id.invalid.reason"),
			t("error.account.notification.id.invalid.recovery"))
		return "", false
	}

	return notificationID, true
}
