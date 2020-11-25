package user

import (
	"errors"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"

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

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	abilities := jsonBody.Abilities
	if len(abilities) == 0 {
		response.NotFoundParameterValue(c, "abilities",
			t(errorMessageKey),
			t("error.admin.user.abilities.not_found.reason"),
			t("error.admin.user.abilities.not_found.recovery"))
		return
	}

	var err error
	if action == "add" {
		if AddAbilitiesToUser == nil {
			err = c.Error(errors.New("undefined function 'AddAbilitiesToUser'"))
		} else {
			err = AddAbilitiesToUser(c, abilities, userID)
		}
	} else if action == "remove" {
		if RemoveAbilitiesFromUser == nil {
			err = c.Error(errors.New("undefined function 'RemoveAbilitiesFromUser'"))
		} else {
			err = RemoveAbilitiesFromUser(c, abilities, userID)
		}
	} else if action == "set" {
		if RemoveAbilitiesFromUser == nil {
			err = c.Error(errors.New("undefined function 'SetAbilitiesToUser'"))
		} else {
			err = SetAbilitiesToUser(c, abilities, userID)
		}
	}

	if err != nil {
		response.SendError(c, err)
		return
	}

	response.OkWithStatus(c, t("status.admin.user.abilities.updated"))
}
