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
func VerifyLatestVersionOfTermsAcceptedByUser(currentTermsVersion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		storage := terms.StorageOf(c)
		if storage == nil {
			c.Error(errors.New("undefined storage 'Terms'"))
			return
		}

		userID := ctxUser.GetID(c)
		if len(userID) == 0 {
			return
		}

		latestVersionOfTermsAcceptedByUser, err := storage.GetLatestVersionOfTermsAcceptedByUser(c, userID)
		if err != nil {
			response.InternalServerError(c)
			return
		}
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
