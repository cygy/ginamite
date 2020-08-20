package cookie

import (
	"net/http"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/html"

	"github.com/gin-gonic/gin"
)

// Create creates a new cookie.
func Create(c *gin.Context, value string) {
	CreateAndRedirect(c, value, "")
}

// CreateAndRedirect creates a new cookie.
func CreateAndRedirect(c *gin.Context, value, redirect string) {
	c.SetCookie(config.Main.Cookie.Name, value, config.Main.Cookie.TTL, config.Main.Cookie.Path, config.Main.Cookie.Domain, config.Main.Cookie.Secure, config.Main.Cookie.HTTPOnly)

	if len(redirect) > 0 {
		locale := context.GetLocale(c)
		c.Redirect(http.StatusFound, html.Main.URL(redirect, locale))
	}
}

// Invalidate invalidates a cookie
func Invalidate(c *gin.Context) {
	InvalidateAndRedirect(c, "")
}

// InvalidateAndRedirect invalidates a cookie
func InvalidateAndRedirect(c *gin.Context, redirect string) {
	c.SetCookie(config.Main.Cookie.Name, "", -1, config.Main.Cookie.Path, config.Main.Cookie.Domain, config.Main.Cookie.Secure, config.Main.Cookie.HTTPOnly)

	if len(redirect) > 0 {
		locale := context.GetLocale(c)
		c.Redirect(http.StatusFound, html.Main.URL(redirect, locale))
	}
}

// GetValue returns the value of a cookie
func GetValue(c *gin.Context) (string, error) {
	return c.Cookie(config.Main.Cookie.Name)
}
