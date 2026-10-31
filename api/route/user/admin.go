// Package user provides the administration handlers to manage the users.
package user

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/common/ctxqueue"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/api/route/account"
	"github.com/cygy/ginamite/common"
	"github.com/cygy/ginamite/common/ctxconfig"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

// GetAdminEnabledUsers returns the enabled users.
func GetAdminEnabledUsers(c *gin.Context) {
	if FunctionsOf(c).GetEnabledUsersForAdmin == nil {
		c.Error(errors.New("undefined function 'GetEnabledUsersForAdmin'"))
		response.InternalServerError(c)
		return
	}

	users, err := FunctionsOf(c).GetEnabledUsersForAdmin(c)
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
	if FunctionsOf(c).GetDisabledUsersForAdmin == nil {
		c.Error(errors.New("undefined function 'GetDisabledUsersForAdmin'"))
		response.InternalServerError(c)
		return
	}

	users, err := FunctionsOf(c).GetDisabledUsersForAdmin(c)
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
	t := localization.Translate(locale)

	// Password is mandatory.
	newPassword := jsonBody.Password
	if len(newPassword) == 0 {
		response.NotFoundParameterValue(c, "password",
			t(errorMessageKey),
			t("error.admin.user.password.update.password.not_found.reason"),
			t("error.admin.user.password.update.password.not_found.recovery"))
		return
	}

	var err error
	if FunctionsOf(c).UpdatePasswordByAdmin == nil {
		err = c.Error(errors.New("undefined function 'UpdatePasswordByAdmin'"))
	} else {
		userID := getUserID(c)
		err = FunctionsOf(c).UpdatePasswordByAdmin(c, userID, jsonBody.Password)
	}

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
	t := localization.Translate(locale)

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

	var err error
	if FunctionsOf(c).UpdateInformationsByAdmin == nil {
		err = c.Error(errors.New("undefined function 'UpdateInformationsByAdmin'"))
	} else {
		userID := getUserID(c)
		err = FunctionsOf(c).UpdateInformationsByAdmin(c, userID, jsonBody.FirstName, jsonBody.LastName, newEmailAddress, jsonBody.ValidEmailAddress)
	}

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
	t := localization.Translate(locale)

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
	if !ctxconfig.Get(c).SupportsLocale(newLocale) {
		response.UnsupportedParameterLocale(c, locale, newLocale, "locale", ctxconfig.Get(c).SupportedLocales, errorMessageKey)
		return
	}

	var err error
	if FunctionsOf(c).UpdateDescriptionByAdmin == nil {
		err = c.Error(errors.New("undefined function 'UpdateDescriptionByAdmin'"))
	} else {
		userID := getUserID(c)
		err = FunctionsOf(c).UpdateDescriptionByAdmin(c, userID, newLocale, jsonBody.Description)
	}

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
	t := localization.Translate(locale)

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

	var err error
	if FunctionsOf(c).UpdateSocialByAdmin == nil {
		err = c.Error(errors.New("undefined function 'UpdateSocialByAdmin'"))
	} else {
		err = FunctionsOf(c).UpdateSocialByAdmin(c, userID, facebook, twitter, instagram)
	}

	if err != nil {
		response.InternalServerError(c)
		return
	}

	ctxqueue.Get(c).UpdateUserSocialNetworks(queue.TaskUpdateUserSocialNetworks{
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
	t := localization.Translate(locale)

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
	if !ctxconfig.Get(c).SupportsLocale(newLocale) {
		response.UnsupportedParameterLocale(c, locale, newLocale, "locale", ctxconfig.Get(c).SupportedLocales, errorMessageKey)
		return
	}

	// Timezone not supported.
	if !timezone.Main.SupportsTimezone(newTimezone) {
		response.InvalidParameterValue(c, "timezone",
			t(errorMessageKey),
			t("error.admin.user.settings.update.timezone.not_supported.reason"),
			t("error.admin.user.settings.update.timezone.not_supported.recovery"),
		)
		return
	}

	var err error
	if FunctionsOf(c).UpdateSettingsByAdmin == nil {
		err = c.Error(errors.New("undefined function 'UpdateSettingsByAdmin'"))
	} else {
		userID := getUserID(c)
		err = FunctionsOf(c).UpdateSettingsByAdmin(c, userID, newLocale, newTimezone)
	}

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
	t := localization.Translate(locale)

	notificationsToSave := jsonBody.Notifications
	for _, notification := range notificationsToSave {
		if !notifications.IsNotificationSupported(notification) {
			response.InvalidParameterValue(c, "notifications",
				t(errorMessageKey),
				t("error.admin.user.notifications.update.invalid.reason", localization.H{"Value": notification}),
				t("error.admin.user.notifications.update.invalid.recovery", localization.H{"Value": strings.Join(notifications.SupportedNotifications(), ",")}))
			return
		}
	}

	var err error
	if FunctionsOf(c).UpdateNotificationsByAdmin == nil {
		err = c.Error(errors.New("undefined function 'UpdateNotificationsByAdmin'"))
	} else {
		userID := getUserID(c)
		err = FunctionsOf(c).UpdateNotificationsByAdmin(c, userID, notificationsToSave)
	}

	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.admin.user.notifications.updated"))
}

// GetAdminUserTokens returns the user's auth tokens.
func GetAdminUserTokens(c *gin.Context) {
	if FunctionsOf(c).GetTokensForAdmin == nil {
		c.Error(errors.New("undefined function 'GetTokensForAdmin'"))
		response.InternalServerError(c)
		return
	}

	userID := getUserID(c)

	tokens, err := FunctionsOf(c).GetTokensForAdmin(c, userID)
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
	if FunctionsOf(c).DeleteTokenByIDByAdmin == nil {
		c.Error(errors.New("undefined function 'DeleteTokenByIDByAdmin'"))
		response.InternalServerError(c)
		return
	}

	if account.FunctionsOf(c).GetTokenDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetTokenDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	tokenID := getTokenID(c)

	token, err := account.FunctionsOf(c).GetTokenDetailsByID(c, tokenID)
	if err != nil && !commonErrors.IsNotFound(err) {
		response.InternalServerError(c)
		return
	}

	// If the token does exist, delete it.
	if len(token.UserID) > 0 {
		if err := FunctionsOf(c).DeleteTokenByIDByAdmin(c, tokenID); err != nil && !commonErrors.IsNotFound(err) {
			response.InternalServerError(c)
			return
		}

		ctxqueue.Get(c).InvalidAuthToken(tokenID, token.ExpirationDate)
	}

	locale := ctxLocale.Get(c)
	response.OkWithStatus(c, localization.Translate(locale)("status.admin.user.token.deleted"))
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

	ctxqueue.Get(c).DeleteUser(queue.TaskDeleteUser{
		UserID: userID,
	})

	// Send a mail to warn the user.
	ctxqueue.Get(c).SendMail(queue.MessageMailDeleteAccountConfirmed, queue.MailDeleteAccountConfirmed{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
	})

	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

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

	ctxqueue.Get(c).DisableUser(queue.TaskDisableUser{
		UserID: userID,
	})

	// Send a mail to warn the user.
	ctxqueue.Get(c).SendMail(queue.MessageMailDisableAccountConfirmed, queue.MailDisableAccountConfirmed{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
	})

	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

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

	ctxqueue.Get(c).EnableUser(queue.TaskEnableUser{
		UserID: user.ID,
	})

	// Send a mail to warn the user.
	ctxqueue.Get(c).SendMail(queue.MessageMailEnableAccountConfirmed, queue.MailEnableAccountConfirmed{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
	})

	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

	response.OkWithStatus(c, t("status.admin.account.enabled"))
}
