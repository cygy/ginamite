// Package response renders the error pages of the web server, and provides the handlers of the API errors.
package response

import (
	"net/http"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"

	"github.com/gin-gonic/gin"
)

// InternalServerError renders the page "error.internal_server" with a HTTP code 500, displaying the ID of the request.
func InternalServerError(c *gin.Context) {
	page := html.Main.PageByKey(c, "error.internal_server")
	if page == nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	requestID := c.Request.Header.Get(log.RequestIDHeaderName)
	locale := context.GetLocale(c)
	page.SetVar("ID", localization.Translate(locale)("tpl.error.internal_server.id", localization.H{"ErrorID": requestID}))

	page.RenderError(c)
}

// InvalidAuthentication renders the page "error.authentication.invalid" with a HTTP code 401.
func InvalidAuthentication(c *gin.Context) {
	renderUnauthorized(c, "error.authentication.invalid")
}

// ExpiredAuthentication renders the page "error.authentication.expired" with a HTTP code 401.
func ExpiredAuthentication(c *gin.Context) {
	renderUnauthorized(c, "error.authentication.expired")
}

// RevokedAuthentication renders the page "error.authentication.revoked" with a HTTP code 401.
func RevokedAuthentication(c *gin.Context) {
	renderUnauthorized(c, "error.authentication.revoked")
}

// Unauthorized renders the page "error.unauthorized" with a HTTP code 401.
func Unauthorized(c *gin.Context) {
	renderUnauthorized(c, "error.unauthorized")
}

// NotFound renders the page "error.not_found" with a HTTP code 404.
func NotFound(c *gin.Context) {
	page := html.Main.PageByKey(c, "error.not_found")
	if page == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	page.RenderNotFound(c)
}

// APIError renders the page with the error of the API.
func APIError(c *gin.Context, page *html.Page, err response.Error) {
	page.SetError(err)
	page.RenderOk(c)
}

// renderUnauthorized renders the page key with a HTTP code 401, or only the code if the page does not exist.
func renderUnauthorized(c *gin.Context, key string) {
	page := html.Main.PageByKey(c, key)
	if page == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	page.RenderUnauthorized(c)
}
