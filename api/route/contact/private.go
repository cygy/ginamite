package contact

import (
	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

func checkContactID(c *gin.Context, messageKey string) (string, bool) {
	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	contactID := c.Param("id")
	if len(contactID) == 0 {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.admin.contact.id.not_found.reason"),
			t("error.admin.contact.id.not_found.recovery"))
		return "", false
	}

	if !bson.IsObjectIdHex(contactID) {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.admin.contact.id.invalid.reason"),
			t("error.admin.contact.id.invalid.recovery"))
		return "", false
	}

	return contactID, true
}
