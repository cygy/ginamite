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

// authenticationErrorPages maps the codes of the authentication errors to their pages.
// The authentication cookie is deleted unless the page is Unauthorized: the token is still good, it is the rights
// which are insufficient.
var authenticationErrorPages = map[int]func(c *gin.Context){
	authErrors.InvalidAuthorizationToken: func(c *gin.Context) {
		cookie.Invalidate(c)
		InvalidAuthentication(c)
	},
	authErrors.ExpiredAuthorizationToken: func(c *gin.Context) {
		cookie.Invalidate(c)
		ExpiredAuthentication(c)
	},
	authErrors.RevokedAuthorizationToken: func(c *gin.Context) {
		cookie.Invalidate(c)
		RevokedAuthentication(c)
	},
	authErrors.InsufficientRights: Unauthorized,
}

// HandleInternalServerError returns a response handler rendering the page of the internal server errors,
// when the API did not respond or answered with a HTTP code 500.
func HandleInternalServerError() api.ResponseHandlerFunc {
	return func(c *gin.Context, statusCode int, requestError error, _ response.Error) bool {
		if requestError != nil || statusCode == http.StatusInternalServerError {
			InternalServerError(c)
			c.Abort()
			return true
		}

		return false
	}
}

// HandleInvalidAuthenticationError returns a response handler rendering the page of the authentication errors
// of the API: invalid, expired or revoked token, and insufficient rights.
func HandleInvalidAuthenticationError() api.ResponseHandlerFunc {
	return func(c *gin.Context, statusCode int, _ error, apiError response.Error) bool {
		if statusCode != http.StatusUnauthorized || apiError.Domain != authErrors.Domain {
			return false
		}

		renderPage, ok := authenticationErrorPages[apiError.Code]
		if !ok {
			return false
		}

		renderPage(c)
		c.Abort()

		return true
	}
}

// HandleLatestTermsVersionMustBeAcceptedError returns a response handler redirecting to the page of the terms
// when the API requires the user to accept their latest version.
func HandleLatestTermsVersionMustBeAcceptedError() api.ResponseHandlerFunc {
	return func(c *gin.Context, statusCode int, _ error, apiError response.Error) bool {
		if statusCode == http.StatusPreconditionFailed && apiError.Domain == termsErrors.Domain && apiError.Code == termsErrors.MustAcceptLatestTerms {
			locale := context.GetLocale(c)
			c.Redirect(http.StatusFound, html.Of(c).URL("account.settings.terms", locale))
			c.Abort()
			return true
		}

		return false
	}
}

// HandleAPIError returns a response handler rendering the page with the error of the API, if there is one.
func HandleAPIError(page *html.Page) api.ResponseHandlerFunc {
	return func(c *gin.Context, _ int, _ error, apiError response.Error) bool {
		if !apiError.IsNil() {
			APIError(c, page, apiError)
			c.Abort()
			return true
		}

		return false
	}
}
