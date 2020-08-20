package context

import (
	"errors"

	"github.com/cygy/ginamite/api/response"
	termsError "github.com/cygy/ginamite/common/errors/terms"
	"github.com/cygy/ginamite/common/localization"
	r "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// GetLatestVersionOfTermsAcceptedByUserFunc : returns the latest version of the terms accepted by the user.
type GetLatestVersionOfTermsAcceptedByUserFunc func(c *gin.Context, userID string) string

// VerifyLatestVersionOfTermsAcceptedByUser : checks and saves the authenticated user to the context.
func VerifyLatestVersionOfTermsAcceptedByUser(currentTermsVersion string, getLatestVersionOfTermsAcceptedByUser GetLatestVersionOfTermsAcceptedByUserFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if getLatestVersionOfTermsAcceptedByUser == nil {
			c.Error(errors.New("undefined function 'GetLatestVersionOfTermsAcceptedByUser'"))
			return
		}

		userID := GetUserID(c)
		if len(userID) == 0 {
			return
		}

		if currentTermsVersion != getLatestVersionOfTermsAcceptedByUser(c, userID) {
			locale := GetLocale(c)
			t := localization.Translate(locale)

			response.PreconditionFailedWithError(c, r.Error{
				Domain:   termsError.Domain,
				Code:     termsError.MustAcceptLatestTerms,
				Message:  t("error.request.terms.must_accept_latest.message"),
				Reason:   t("error.request.terms.must_accept_latest.reason"),
				Recovery: t("error.request.terms.must_accept_latest.recovery"),
			})
		}
	}
}
