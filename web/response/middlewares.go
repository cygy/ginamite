package response

import (
	"net/http"

	authErrors "github.com/cygy/ginamite/common/errors/auth"
	termsErrors "github.com/cygy/ginamite/common/errors/terms"
	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/cookie"
	"github.com/cygy/ginamite/web/html"

	"github.com/gin-gonic/gin"
)

// HandleInternalServerError : returns a API response handler.
func HandleInternalServerError() api.ResponseHandlerFunc {
	return func(c *gin.Context, statusCode int, requestError error, apiError response.Error) bool {
		if requestError != nil || statusCode == http.StatusInternalServerError {
			InternalServerError(c)
			c.Abort()
			return true
		}

		return false
	}
}

// HandleInvalidAuthenticationError : returns a API response handler.
func HandleInvalidAuthenticationError() api.ResponseHandlerFunc {
	return func(c *gin.Context, statusCode int, requestError error, apiError response.Error) bool {
		if statusCode != http.StatusUnauthorized || apiError.Domain != authErrors.Domain {
			return false
		}

		handled := false

		switch apiError.Code {
		case authErrors.InvalidAuthorizationToken:
			cookie.Invalidate(c)
			InvalidAuthentication(c)
			c.Abort()
			handled = true
		case authErrors.ExpiredAuthorizationToken:
			cookie.Invalidate(c)
			ExpiredAuthentication(c)
			c.Abort()
			handled = true
		case authErrors.RevokedAuthorizationToken:
			cookie.Invalidate(c)
			RevokedAuthentication(c)
			c.Abort()
			handled = true
		case authErrors.InsufficientRights:
			Unauthorized(c)
			c.Abort()
			handled = true
		}

		return handled
	}
}

// HandleLatestTermsVersionMustBeAcceptedError : returns a API response handler.
func HandleLatestTermsVersionMustBeAcceptedError() api.ResponseHandlerFunc {
	return func(c *gin.Context, statusCode int, requestError error, apiError response.Error) bool {
		if statusCode == http.StatusPreconditionFailed && apiError.Domain == termsErrors.Domain && apiError.Code == termsErrors.MustAcceptLatestTerms {
			locale := context.GetLocale(c)
			c.Redirect(http.StatusPreconditionFailed, html.Main.URL("account.settings.terms", locale))
			c.Abort()
			return true
		}

		return false
	}
}

// HandleAPIError : returns a API error handler.
func HandleAPIError(page *html.Page) api.ResponseHandlerFunc {
	return func(c *gin.Context, statusCode int, requestError error, apiError response.Error) bool {
		if !apiError.IsNil() {
			APIError(c, page, apiError)
			c.Abort()
			return true
		}

		return false
	}
}
