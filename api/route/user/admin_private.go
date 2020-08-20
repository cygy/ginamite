package user

import (
	"errors"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

func checkUserID(c *gin.Context, messageKey string) (string, bool) {
	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	userID := c.Param("id")
	if len(userID) == 0 {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.admin.user.id.not_found.reason"),
			t("error.admin.user.id.not_found.recovery"))
		return "", false
	}

	if !bson.IsObjectIdHex(userID) {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.admin.user.id.invalid.reason"),
			t("error.admin.user.id.invalid.recovery"))
		return "", false
	}

	return userID, true
}

func checkTokenID(c *gin.Context, messageKey string) (string, bool) {
	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	tokenID := c.Param("id")
	if len(tokenID) == 0 {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.admin.token.id.not_found.reason"),
			t("error.admin.token.id.not_found.recovery"))
		return "", false
	}

	if !bson.IsObjectIdHex(tokenID) {
		response.InvalidRequestParameterWithDetails(c,
			t(messageKey),
			t("error.admin.token.id.invalid.reason"),
			t("error.admin.token.id.invalid.recovery"))
		return "", false
	}

	return tokenID, true
}

func updateAbilities(c *gin.Context, action string) {
	errorMessageKey := "error.admin.user.abilities.update.message"

	userID, ok := checkUserID(c, errorMessageKey)
	if !ok {
		return
	}

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
