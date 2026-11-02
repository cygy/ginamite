package user

import (
	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"

	"github.com/gin-gonic/gin"
)

func getUserID(c *gin.Context) string {
	return c.Param("id")
}

func getTokenID(c *gin.Context) string {
	return c.Param("id")
}

func updateAbilities(c *gin.Context, action string) {
	errorMessageKey := "error.admin.user.abilities.update.message"

	userID := getUserID(c)

	var jsonBody struct {
		Abilities []string `json:"abilities"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	abilities := jsonBody.Abilities
	if len(abilities) == 0 {
		response.NotFoundParameterValue(c, "abilities",
			t(errorMessageKey),
			t("error.admin.user.abilities.not_found.reason"),
			t("error.admin.user.abilities.not_found.recovery"))
		return
	}

	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	var err error
	switch action {
	case "add":
		err = storage.AddAbilitiesToUser(c, userID, abilities)
	case "remove":
		err = storage.RemoveAbilitiesFromUser(c, userID, abilities)
	case "set":
		err = storage.SetAbilitiesToUser(c, userID, abilities)
	}

	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OkWithStatus(c, t("status.admin.user.abilities.updated"))
}
