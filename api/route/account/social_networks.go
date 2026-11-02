package account

import (
	"strings"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

// UpdateSocialNetworks updates the social network profiles of the user.
func UpdateSocialNetworks(c *gin.Context) {
	if StorageOf(c).Accounts == nil {
		c.Error(errNoAccounts)
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Facebook  string `json:"facebook"`
		Twitter   string `json:"twitter"`
		Instagram string `json:"instagram"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	errorMessageKey := "error.account.social.update.message"

	// The facebook profile must begin with a valid url.
	facebook := jsonBody.Facebook
	if len(facebook) > 0 && !strings.HasPrefix(facebook, common.FacebookPrefix) {
		response.InvalidParameterValue(c, "facebook",
			t(errorMessageKey),
			t("error.account.social.facebook.prefix.reason"),
			t("error.account.social.facebook.prefix.recovery", localization.H{"Prefix": common.FacebookPrefix}),
		)
		return
	}

	// The twitter profile must begin with a valid url.
	twitter := jsonBody.Twitter
	if len(twitter) > 0 && !strings.HasPrefix(twitter, common.TwitterPrefix) {
		response.InvalidParameterValue(c, "twitter",
			t(errorMessageKey),
			t("error.account.social.twitter.prefix.reason"),
			t("error.account.social.twitter.prefix.recovery", localization.H{"Prefix": common.TwitterPrefix}),
		)
		return
	}

	// The instagram profile must begin with a valid url.
	instagram := jsonBody.Instagram
	if len(instagram) > 0 && !strings.HasPrefix(instagram, common.InstagramPrefix) {
		response.InvalidParameterValue(c, "instagram",
			t(errorMessageKey),
			t("error.account.social.instagram.prefix.reason"),
			t("error.account.social.instagram.prefix.recovery", localization.H{"Prefix": common.InstagramPrefix}),
		)
		return
	}

	userID := ctxUser.GetID(c)
	if err := StorageOf(c).Accounts.UpdateUserSocialNetworks(c, userID, facebook, twitter, instagram); err != nil {
		response.InternalServerError(c)
		return
	}

	ctxkey.Queue.Get(c).UpdateUserSocialNetworks(queue.TaskUpdateUserSocialNetworks{
		UserID: userID,
	})

	response.OkWithStatus(c, t("status.account.social.updated"))
}
