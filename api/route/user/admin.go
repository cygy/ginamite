// Package user provides the administration handlers to manage the users.
package user

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/common"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

// errNoTokens is logged when the route to delete a token is called without storage of the auth tokens.
var errNoTokens = errors.New("undefined storage of the auth tokens")

// errNoStorage is logged when a route of the administration is called without storage of the users.
var errNoStorage = errors.New("undefined storage of the users")

// GetAdminEnabledUsers returns the enabled users.
func GetAdminEnabledUsers(c *gin.Context) {
	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	users, err := storage.GetEnabledUsersForAdmin(c)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"users": users,
	})
}

// GetAdminDisabledUsers returns the disabled users.
func GetAdminDisabledUsers(c *gin.Context) {
	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	users, err := storage.GetDisabledUsersForAdmin(c)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"users": users,
	})
}

// AdminUpdatePassword updates the password of a user.
func AdminUpdatePassword(c *gin.Context) {
	errorMessageKey := "error.admin.user.password.update.message"

	var jsonBody struct {
		Password string `json:"password"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// Password is mandatory.
	newPassword := jsonBody.Password
	if len(newPassword) == 0 {
		response.NotFoundParameterValue(c, "password",
			t(errorMessageKey),
			t("error.admin.user.password.update.password.not_found.reason"),
			t("error.admin.user.password.update.password.not_found.recovery"))
		return
	}

	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	userID := getUserID(c)
	err := storage.UpdatePasswordByAdmin(c, userID, jsonBody.Password)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.admin.user.password.updated"))
}

// AdminUpdateInformations updates the informations of a user.
func AdminUpdateInformations(c *gin.Context) {
	errorMessageKey := "error.admin.user.informations.update.message"

	var jsonBody struct {
		FirstName         string `json:"first_name"`
		LastName          string `json:"last_name"`
		EmailAddress      string `json:"email_address"`
		ValidEmailAddress bool   `json:"valid_email_address"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// EmailAddress is mandatory.
	newEmailAddress := jsonBody.EmailAddress
	if len(newEmailAddress) == 0 {
		response.NotFoundParameterValue(c, "email_address",
			t(errorMessageKey),
			t("error.admin.user.informations.update.email_address.not_found.reason"),
			t("error.admin.user.informations.update.email_address.not_found.recovery"))
		return
	}

	// Email address must be valid.
	if !strings.Contains(newEmailAddress, "@") {
		response.InvalidParameterValue(c, "email_address",
			t(errorMessageKey),
			t("error.admin.user.informations.update.email_address.invalid.reason"),
			t("error.admin.user.informations.update.email_address.invalid.recovery"),
		)
		return
	}

	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	userID := getUserID(c)
	err := storage.UpdateInformationsByAdmin(c, userID, jsonBody.FirstName, jsonBody.LastName, newEmailAddress, jsonBody.ValidEmailAddress)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.admin.user.informations.updated"))
}

// AdminUpdateDescription updates the description of a user.
func AdminUpdateDescription(c *gin.Context) {
	errorMessageKey := "error.admin.user.description.update.message"

	var jsonBody struct {
		Description string `json:"description"`
		Locale      string `json:"locale"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// Locale is mandatory.
	newLocale := jsonBody.Locale
	if len(newLocale) == 0 {
		response.NotFoundParameterValue(c, "locale",
			t(errorMessageKey),
			t("error.admin.user.description.update.locale.not_found.reason"),
			t("error.admin.user.description.update.locale.not_found.recovery"))
		return
	}

	// Locale not supported.
	if !ctxkey.Config.Get(c).SupportsLocale(newLocale) {
		response.UnsupportedParameterLocale(c, locale, newLocale, "locale", ctxkey.Config.Get(c).SupportedLocales, errorMessageKey)
		return
	}

	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	userID := getUserID(c)
	err := storage.UpdateDescriptionByAdmin(c, userID, newLocale, jsonBody.Description)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.admin.user.description.updated"))
}

// AdminUpdateSocial updates the social networks of a user.
func AdminUpdateSocial(c *gin.Context) {
	errorMessageKey := "error.admin.user.social.update.message"

	var jsonBody struct {
		Facebook  string `json:"facebook"`
		Twitter   string `json:"twitter"`
		Instagram string `json:"instagram"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// The facebook profile must begin with a valid url.
	facebook := jsonBody.Facebook
	if len(facebook) > 0 && !strings.HasPrefix(facebook, common.FacebookPrefix) {
		response.InvalidParameterValue(c, "facebook",
			t(errorMessageKey),
			t("error.admin.user.social.update.facebook.prefix.reason"),
			t("error.admin.user.social.update.facebook.prefix.recovery", localization.H{"Prefix": common.FacebookPrefix}),
		)
		return
	}

	// The twitter profile must begin with a valid url.
	twitter := jsonBody.Twitter
	if len(twitter) > 0 && !strings.HasPrefix(twitter, common.TwitterPrefix) {
		response.InvalidParameterValue(c, "twitter",
			t(errorMessageKey),
			t("error.admin.user.social.update.twitter.prefix.reason"),
			t("error.admin.user.social.update.twitter.prefix.recovery", localization.H{"Prefix": common.TwitterPrefix}),
		)
		return
	}

	// The instagram profile must begin with a valid url.
	instagram := jsonBody.Instagram
	if len(instagram) > 0 && !strings.HasPrefix(instagram, common.InstagramPrefix) {
		response.InvalidParameterValue(c, "instagram",
			t(errorMessageKey),
			t("error.admin.user.social.update.instagram.prefix.reason"),
			t("error.admin.user.social.update.instagram.prefix.recovery", localization.H{"Prefix": common.InstagramPrefix}),
		)
		return
	}

	userID := getUserID(c)

	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	err := storage.UpdateSocialByAdmin(c, userID, facebook, twitter, instagram)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	ctxkey.Queue.Get(c).UpdateUserSocialNetworks(queue.TaskUpdateUserSocialNetworks{
		UserID: userID,
	})

	response.OkWithStatus(c, t("status.admin.user.social.updated"))
}

// AdminUpdateSettings updates the settings of a user.
func AdminUpdateSettings(c *gin.Context) {
	errorMessageKey := "error.admin.user.settings.update.message"

	var jsonBody struct {
		Locale   string `json:"locale"`
		Timezone string `json:"timezone"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// Locale is mandatory.
	newLocale := jsonBody.Locale
	if len(newLocale) == 0 {
		response.NotFoundParameterValue(c, "locale",
			t(errorMessageKey),
			t("error.admin.user.settings.update.locale.not_found.reason"),
			t("error.admin.user.settings.update.locale.not_found.recovery"))
		return
	}

	// Timezone is mandatory.
	newTimezone := jsonBody.Timezone
	if len(newTimezone) == 0 {
		response.NotFoundParameterValue(c, "timezone",
			t(errorMessageKey),
			t("error.admin.user.settings.update.timezone.not_found.reason"),
			t("error.admin.user.settings.update.timezone.not_found.recovery"))
		return
	}

	// Locale not supported.
	if !ctxkey.Config.Get(c).SupportsLocale(newLocale) {
		response.UnsupportedParameterLocale(c, locale, newLocale, "locale", ctxkey.Config.Get(c).SupportedLocales, errorMessageKey)
		return
	}

	// Timezone not supported.
	if !ctxkey.Timezones.Get(c).SupportsTimezone(newTimezone) {
		response.InvalidParameterValue(c, "timezone",
			t(errorMessageKey),
			t("error.admin.user.settings.update.timezone.not_supported.reason"),
			t("error.admin.user.settings.update.timezone.not_supported.recovery"),
		)
		return
	}

	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	userID := getUserID(c)
	err := storage.UpdateSettingsByAdmin(c, userID, newLocale, newTimezone)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.admin.user.settings.updated"))
}

// AdminUpdateNotifications updates the notifications of a user.
func AdminUpdateNotifications(c *gin.Context) {
	errorMessageKey := "error.admin.user.notifications.update.message"

	var jsonBody struct {
		Notifications []string `json:"notifications"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	notificationsToSave := jsonBody.Notifications
	for _, notification := range notificationsToSave {
		if !ctxkey.Notifications.Get(c).IsSupported(notification) {
			response.InvalidParameterValue(c, "notifications",
				t(errorMessageKey),
				t("error.admin.user.notifications.update.invalid.reason", localization.H{"Value": notification}),
				t("error.admin.user.notifications.update.invalid.recovery", localization.H{"Value": strings.Join(ctxkey.Notifications.Get(c).Supported(), ",")}))
			return
		}
	}

	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	userID := getUserID(c)
	err := storage.UpdateNotificationsByAdmin(c, userID, notificationsToSave)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.admin.user.notifications.updated"))
}

// GetAdminUserTokens returns the user's auth tokens.
func GetAdminUserTokens(c *gin.Context) {
	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	userID := getUserID(c)

	tokens, err := storage.GetTokensForAdmin(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"tokens": tokens,
	})
}

// AdminDeleteUserToken deletes a user's auth token.
func AdminDeleteUserToken(c *gin.Context) {
	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	if account.FunctionsOf(c).Tokens == nil {
		c.Error(errNoTokens)
		response.InternalServerError(c)
		return
	}

	tokenID := getTokenID(c)

	token, err := account.FunctionsOf(c).Tokens.GetTokenDetailsByID(c, tokenID)
	if err != nil && !commonErrors.IsNotFound(err) {
		response.InternalServerError(c)
		return
	}

	// If the token does exist, delete it.
	if len(token.UserID) > 0 {
		if err := storage.DeleteTokenByIDByAdmin(c, tokenID); err != nil && !commonErrors.IsNotFound(err) {
			response.InternalServerError(c)
			return
		}

		ctxkey.Queue.Get(c).InvalidAuthToken(tokenID, token.ExpirationDate)
	}

	locale := ctxLocale.Get(c)
	response.OkWithStatus(c, ctxkey.Localization.Get(c).Translate(locale)("status.admin.user.token.deleted"))
}

// SetAbilities set abilities to a user.
func SetAbilities(c *gin.Context) {
	updateAbilities(c, "set")
}

// AddAbilities add abilities to a user.
func AddAbilities(c *gin.Context) {
	updateAbilities(c, "add")
}

// RemoveAbilities remove abilities from a user.
func RemoveAbilities(c *gin.Context) {
	updateAbilities(c, "remove")
}

// AdminDelete deletes a user.
func AdminDelete(c *gin.Context) {
	// Get the userID from the request.
	userID := getUserID(c)

	// Get the user to delete.
	user, err := account.FunctionsOf(c).GetUserDetailsByID(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// Delete the user.
	if err := account.FunctionsOf(c).DeleteUserByID(c, userID); err != nil {
		response.InternalServerError(c)
		return
	}

	ctxkey.Queue.Get(c).DeleteUser(queue.TaskDeleteUser{
		UserID: userID,
	})

	// Send a mail to warn the user.
	ctxkey.Queue.Get(c).SendMail(queue.MessageMailDeleteAccountConfirmed, queue.MailDeleteAccountConfirmed{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
	})

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	response.OkWithStatus(c, t("status.admin.account.deleted"))
}

// AdminDisable disables a user.
func AdminDisable(c *gin.Context) {
	// Get the userID from the request.
	userID := getUserID(c)

	// Disable the user.
	user, err := account.FunctionsOf(c).DisableUserByID(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	ctxkey.Queue.Get(c).DisableUser(queue.TaskDisableUser{
		UserID: userID,
	})

	// Send a mail to warn the user.
	ctxkey.Queue.Get(c).SendMail(queue.MessageMailDisableAccountConfirmed, queue.MailDisableAccountConfirmed{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
	})

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	response.OkWithStatus(c, t("status.admin.account.disabled"))
}

// AdminEnable enables a user.
func AdminEnable(c *gin.Context) {
	// Get the userID from the request.
	disabledUserID := getUserID(c)

	// Enable the user.
	user, err := account.FunctionsOf(c).EnableUserByID(c, disabledUserID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	ctxkey.Queue.Get(c).EnableUser(queue.TaskEnableUser{
		UserID: user.ID,
	})

	// Send a mail to warn the user.
	ctxkey.Queue.Get(c).SendMail(queue.MessageMailEnableAccountConfirmed, queue.MailEnableAccountConfirmed{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
	})

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	response.OkWithStatus(c, t("status.admin.account.enabled"))
}
