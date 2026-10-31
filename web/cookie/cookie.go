// Package cookie manages the authentication cookie of the web server, configured by the Cookie settings of the configuration in the context, see ctxconfig.
package cookie

import (
	"net/http"

	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"

	"github.com/gin-gonic/gin"
)

// Create saves value in the authentication cookie.
func Create(c *gin.Context, value string) {
	CreateAndRedirect(c, value, "")
}

// CreateAndRedirect saves value in the authentication cookie, then redirects to the route key redirect in the locale
// of the request, if redirect is not empty.
func CreateAndRedirect(c *gin.Context, value, redirect string) {
	setCookie(c, value, ctxconfig.Get(c).Cookie.TTL)
	redirectTo(c, redirect)
}

// Invalidate deletes the authentication cookie.
func Invalidate(c *gin.Context) {
	InvalidateAndRedirect(c, "")
}

// InvalidateAndRedirect deletes the authentication cookie, then redirects to the route key redirect in the locale
// of the request, if redirect is not empty.
func InvalidateAndRedirect(c *gin.Context, redirect string) {
	setCookie(c, "", -1)
	redirectTo(c, redirect)
}

// GetValue returns the value of the authentication cookie.
func GetValue(c *gin.Context) (string, error) {
	return c.Cookie(ctxconfig.Get(c).Cookie.Name)
}

// setCookie sends the authentication cookie with the maximum age in seconds, which deletes it if negative.
func setCookie(c *gin.Context, value string, maxAge int) {
	cookie := ctxconfig.Get(c).Cookie
	c.SetCookie(cookie.Name, value, maxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HTTPOnly)
}

// redirectTo sends the response redirecting to the route key in the locale of the request, if key is not empty.
func redirectTo(c *gin.Context, key string) {
	if len(key) == 0 {
		return
	}

	c.Redirect(http.StatusFound, html.Main.URL(key, context.GetLocale(c)))
	c.Writer.WriteHeaderNow()
}
