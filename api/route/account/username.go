package account

import (
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/gin-gonic/gin"
)

// GetAvailability returns true if the username is available.
func GetAvailability(c *gin.Context) {
	var jsonBody struct {
		Username string `json:"username"`
	}
	request.DecodeBody(c, &jsonBody)

	username := jsonBody.Username

	if !CheckUsernameParameter(c, username, "username") {
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	if FunctionsOf(c).DoesAccountWithUsernameExist != nil {
		exists, err := FunctionsOf(c).DoesAccountWithUsernameExist(c, username)
		if exists || err != nil {
			response.Ok(c, Availability{
				Available: false,
				Message:   t("status.username.unavailable"),
			})
			return
		}
	}

	response.Ok(c, Availability{
		Available: true,
		Message:   t("status.username.available"),
	})
}

// Availability describes the availability of a username.
type Availability struct {
	Available bool   `json:"available"`
	Message   string `json:"message"`
}
