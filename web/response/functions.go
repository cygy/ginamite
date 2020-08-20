package response

import (
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"

	"github.com/gin-gonic/gin"
)

// InternalServerError : returns a HTTP 5OO error.
func InternalServerError(c *gin.Context) {
	page := html.Main.PageByKey(c, "error.internal_server")

	requestID := c.Request.Header.Get(log.RequestIDHeaderName)
	locale := context.GetLocale(c)
	page.SetVar("ID", localization.Translate(locale)("tpl.error.internal_server.id", localization.H{"ErrorID": requestID}))

	page.RenderError(c)
}

// InvalidAuthentication : returns a HTTP 401 error.
func InvalidAuthentication(c *gin.Context) {
	html.Main.PageByKey(c, "error.authentication.invalid").RenderUnauthorized(c)
}

// ExpiredAuthentication : returns a HTTP 401 error.
func ExpiredAuthentication(c *gin.Context) {
	html.Main.PageByKey(c, "error.authentication.expired").RenderUnauthorized(c)
}

// RevokedAuthentication : returns a HTTP 401 error.
func RevokedAuthentication(c *gin.Context) {
	html.Main.PageByKey(c, "error.authentication.revoked").RenderUnauthorized(c)
}

// Unauthorized : returns a HTTP 401 error.
func Unauthorized(c *gin.Context) {
	html.Main.PageByKey(c, "error.unauthorized").RenderUnauthorized(c)
}

// NotFound : returns the page "Not Found"
func NotFound(c *gin.Context) {
	html.Main.PageByKey(c, "error.not_found").RenderNotFound(c)
}

// APIError : handles a API error.
func APIError(c *gin.Context, page *html.Page, err response.Error) {
	page.SetError(err)
	page.RenderOk(c)
}
