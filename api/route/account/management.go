package account

import (
	"github.com/cygy/ginamite/common/ctxkey"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"

	"github.com/gin-gonic/gin"
)

// GetInfos returns the details of an account.
func GetInfos(c *gin.Context) {
	if StorageOf(c).Accounts == nil {
		c.Error(errNoAccounts)
		response.InternalServerError(c)
		return
	}

	userID := ctxUser.GetID(c)
	user, err := StorageOf(c).Accounts.GetOwnedAccountDetails(c, userID)
	if err != nil || user == nil {
		if err != nil {
			c.Error(err)
		}
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"profile": user,
	})
}

// UpdateInfos updated the details of an account.
func UpdateInfos(c *gin.Context) {
	if StorageOf(c).Accounts == nil {
		c.Error(errNoAccounts)
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if !request.DecodeBody(c, &jsonBody) {
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

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
	if err := StorageOf(c).Accounts.UpdatedAccountInfos(c, userID, firstName, lastName); err != nil {
		response.InternalServerError(c)
		return
	}

	// The auth token is updated too with the latest properties of the user.
	ctxAuthToken.Refresh(c)

	response.OkWithStatus(c, t("status.account.infos.updated"))
}

// UpdateSettings updated the settings of an account.
func UpdateSettings(c *gin.Context) {
	if StorageOf(c).Accounts == nil {
		c.Error(errNoAccounts)
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Locale   string `json:"locale"`
		Timezone string `json:"timezone"`
	}
	if !request.DecodeBody(c, &jsonBody) {
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

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
	if !ctxkey.Config.Get(c).SupportsLocale(accountLocale) {
		response.UnsupportedParameterLocale(c, locale, accountLocale, "locale", ctxkey.Config.Get(c).SupportedLocales, errorMessageKey)
		return
	}

	// Timezone not supported.
	if !ctxkey.Timezones.Get(c).SupportsTimezone(accountTimezone) {
		response.InvalidParameterValue(c, "timezone",
			t(errorMessageKey),
			t("error.account.settings.update.timezone.not_supported.reason"),
			t("error.account.settings.update.timezone.not_supported.recovery"),
		)
		return
	}

	// Update the settings of the user.
	userID := ctxUser.GetID(c)
	if err := StorageOf(c).Accounts.UpdatedAccountSettings(c, userID, accountLocale, accountTimezone); err != nil {
		response.InternalServerError(c)
		return
	}

	// The auth token is updated too with the latest properties of the user.
	ctxAuthToken.Refresh(c)

	response.OkWithStatus(c, t("status.account.settings.updated"))
}
