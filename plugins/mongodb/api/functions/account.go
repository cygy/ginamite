// Package functions implements the functions of the API that read and write the data in MongoDB.
// They are meant to be registered in the API routes by the plugin, see the Plugin type of the parent package.
package functions

import (
	"context"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/cygy/ginamite/common/localization"
	pluginContext "github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/api/ctxsettings"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"
	pluginQueue "github.com/cygy/ginamite/plugins/mongodb/queue"

	"github.com/gin-gonic/gin"
)

// GetOwnedAccountDetails returns the details of the user's account.
func (a *Accounts) GetOwnedAccountDetails(ctx context.Context, userID string) (any, error) {
	user := &model.User{}
	if err := user.GetByID(ctx, userID, a.database()); err != nil {
		return nil, err
	}

	return model.NewOwnUser(user), nil
}

// UpdatedAccountInfos updates the details of a user's account.
func (a *Accounts) UpdatedAccountInfos(ctx context.Context, userID, firstName, lastName string) error {
	user := &model.User{}
	if err := user.GetByID(ctx, userID, a.database()); err != nil {
		return err
	}

	return user.UpdateInfos(ctx, firstName, lastName, a.database())
}

// UpdatedAccountSettings updates the settings of a user's account.
func (a *Accounts) UpdatedAccountSettings(ctx context.Context, userID, locale, timezone string) error {
	user := &model.User{}
	if err := user.GetByID(ctx, userID, a.database()); err != nil {
		return err
	}

	return user.UpdateSettings(ctx, locale, timezone, a.database())
}

// UpdateAccountPublicProfile updates the public profile of an account.
func UpdateAccountPublicProfile(c *gin.Context) {
	settings := ctxsettings.GetSettings(c)
	errorMessageKey := "error.account.public.profile.update.message"

	var jsonBody struct {
		Locale      string `json:"locale"`
		Description string `json:"description"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// The locale must be present and supported.
	usedLocale := jsonBody.Locale
	if !ctxkey.Config.Get(c).SupportsLocale(usedLocale) {
		response.UnsupportedParameterLocale(c, locale, usedLocale, "locale", ctxkey.Config.Get(c).SupportedLocales, errorMessageKey)
		return
	}

	// Description must not be too long.
	description := jsonBody.Description
	if len(description) > int(settings.UserMaxDescriptionLength) {
		response.InvalidParameterValue(c, "description",
			t(errorMessageKey),
			t("error.account.public.profile.description.too_long.reason", localization.H{"Count": settings.UserMaxDescriptionLength}),
			t("error.account.public.profile.description.too_long.recovery", localization.H{"Count": settings.UserMaxDescriptionLength}),
		)
		return
	}

	user := pluginContext.GetUser(c)
	mongoSession := session.Get(c)

	if err := user.UpdatePublicProfile(c, usedLocale, description, mongoSession); err != nil {
		response.InternalServerError(c)
		return
	}

	ctxkey.Queue.Get(c).CreateTask(pluginQueue.MessageTaskUpdateUserPublicProfile, pluginQueue.TaskUpdateUserPublicProfile{
		UserID: user.ID.Hex(),
	})

	response.OkWithStatus(c, t("status.account.public.profile.updated"))
}
