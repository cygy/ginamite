package account

import (
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/gin-gonic/gin"
)

// checkNotificationID returns the identifier of the notification in the URL, or writes an error response and returns
// false if it is missing or not a valid ID, see request.IsValidID.
func checkNotificationID(c *gin.Context, messageKey string) (string, bool) {
	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	notificationID := c.Param("id")
	if len(notificationID) == 0 {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.account.notification.id.not_found.reason"),
			t("error.account.notification.id.not_found.recovery"))
		return "", false
	}

	if !request.IsValidID(c, notificationID) {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.account.notification.id.invalid.reason"),
			t("error.account.notification.id.invalid.recovery"))
		return "", false
	}

	return notificationID, true
}
