package account

import (
	"errors"

	"github.com/cygy/ginamite/common/ctxlocalization"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

// GetInfos returns the details of an account.
func GetInfos(c *gin.Context) {
	if FunctionsOf(c).GetOwnedAccountDetails == nil {
		c.Error(errors.New("undefined function 'GetOwnedAccountDetails'"))
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)
	user := FunctionsOf(c).GetOwnedAccountDetails(c, userID)
	if user == nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"profile": user,
	})
}

// UpdateInfos updated the details of an account.
func UpdateInfos(c *gin.Context) {
	if FunctionsOf(c).UpdatedAccountInfos == nil {
		c.Error(errors.New("undefined function 'UpdatedAccountInfos'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxlocalization.Get(c).Translate(locale)

	errorMessageKey := "error.account.infos.update.unable.message"

	// FirstName is mandatory.
	firstName := jsonBody.FirstName
	if len(firstName) == 0 {
		response.NotFoundParameterValue(c, "first_name",
			t(errorMessageKey),
			t("error.account.infos.update.first_name.not_found.reason"),
			t("error.account.infos.update.first_name.not_found.recovery"),
		)
		return
	}

	// LastName is mandatory.
	lastName := jsonBody.LastName
	if len(lastName) == 0 {
		response.NotFoundParameterValue(c, "last_name",
			t(errorMessageKey),
			t("error.account.infos.update.last_name.not_found.reason"),
			t("error.account.infos.update.last_name.not_found.recovery"),
		)
		return
	}

	// Update the infos of the user.
	userID := ctxUser.GetID(c)
	if err := FunctionsOf(c).UpdatedAccountInfos(c, userID, firstName, lastName); err != nil {
		response.InternalServerError(c)
		return
	}

	// The auth token is updated too with the latest properties of the user.
	ctxAuthToken.Refresh(c)

	response.OkWithStatus(c, t("status.account.infos.updated"))
}

// UpdateSettings updated the settings of an account.
func UpdateSettings(c *gin.Context) {
	if FunctionsOf(c).UpdatedAccountSettings == nil {
		c.Error(errors.New("undefined function 'UpdatedAccountSettings'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Locale   string `json:"locale"`
		Timezone string `json:"timezone"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxlocalization.Get(c).Translate(locale)

	errorMessageKey := "error.account.settings.update.unable.message"

	// Locale is mandatory.
	accountLocale := jsonBody.Locale
	if len(accountLocale) == 0 {
		response.NotFoundParameterValue(c, "locale",
			t(errorMessageKey),
			t("error.account.settings.update.locale.not_found.reason"),
			t("error.account.settings.update.locale.not_found.recovery"),
		)
		return
	}

	// Timezone is mandatory.
	accountTimezone := jsonBody.Timezone
	if len(accountTimezone) == 0 {
		response.NotFoundParameterValue(c, "timezone",
			t(errorMessageKey),
			t("error.account.settings.update.timezone.not_found.reason"),
			t("error.account.settings.update.timezone.not_found.recovery"),
		)
		return
	}

	// Locale not supported.
	if !ctxconfig.Get(c).SupportsLocale(accountLocale) {
		response.UnsupportedParameterLocale(c, locale, accountLocale, "locale", ctxconfig.Get(c).SupportedLocales, errorMessageKey)
		return
	}

	// Timezone not supported.
	if !timezone.Main.SupportsTimezone(accountTimezone) {
		response.InvalidParameterValue(c, "timezone",
			t(errorMessageKey),
			t("error.account.settings.update.timezone.not_supported.reason"),
			t("error.account.settings.update.timezone.not_supported.recovery"),
		)
		return
	}

	// Update the settings of the user.
	userID := ctxUser.GetID(c)
	if err := FunctionsOf(c).UpdatedAccountSettings(c, userID, accountLocale, accountTimezone); err != nil {
		response.InternalServerError(c)
		return
	}

	// The auth token is updated too with the latest properties of the user.
	ctxAuthToken.Refresh(c)

	response.OkWithStatus(c, t("status.account.settings.updated"))
}
