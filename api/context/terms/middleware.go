// Package terms provides the middleware verifying that a user accepted the latest version of the terms.
package terms

import (
	"errors"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/api/route/terms"
	termsError "github.com/cygy/ginamite/common/errors/terms"
	r "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// VerifyLatestVersionOfTermsAcceptedByUser checks and saves the authenticated user to the context.
func VerifyLatestVersionOfTermsAcceptedByUser(currentTermsVersion string, getLatestVersionOfTermsAcceptedByUser terms.GetLatestVersionOfTermsAcceptedByUserFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if getLatestVersionOfTermsAcceptedByUser == nil {
			c.Error(errors.New("undefined function 'GetLatestVersionOfTermsAcceptedByUser'"))
			return
		}

		userID := ctxUser.GetID(c)
		if len(userID) == 0 {
			return
		}

		latestVersionOfTermsAcceptedByUser := getLatestVersionOfTermsAcceptedByUser(c, userID)
		if currentTermsVersion == latestVersionOfTermsAcceptedByUser {
			return
		}

		locale := ctxLocale.Get(c)
		t := ctxkey.Localization.Get(c).Translate(locale)

		response.PreconditionFailedWithError(c, r.Error{
			Domain:   termsError.Domain,
			Code:     termsError.MustAcceptLatestTerms,
			Message:  t("error.request.terms.must_accept_latest.message"),
			Reason:   t("error.request.terms.must_accept_latest.reason"),
			Recovery: t("error.request.terms.must_accept_latest.recovery"),
		})
	}
}
