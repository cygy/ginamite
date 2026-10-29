package account

import (
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
)

// GetAvailability : returns true if the username is available.
func GetAvailability(c *gin.Context) {
	var jsonBody struct {
		Username string `json:"username"`
	}
	request.DecodeBody(c, &jsonBody)

	username := jsonBody.Username

	if !CheckUsernameParameter(c, username, "username") {
		return
	}

	locale := ctxLocale.GetLocale(c)
	t := localization.Translate(locale)

	if DoesAccountWithUsernameExist != nil {
		exists, err := DoesAccountWithUsernameExist(c, username)
		if !exists || err != nil {
			response.Ok(c, Availability{
				Available: false,
				Message:   t("status.username.unavailable"),
			})
		}
	}

	response.Ok(c, Availability{
		Available: true,
		Message:   t("status.username.available"),
	})
}
