package context

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SaveLocale returns a middleware saving the locale of the target page to the context.
func SaveLocale(contextLocale string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(localeKey, contextLocale)
	}
}

// SaveUnknownLocale returns a middleware saving to the context the first supported locale found in the path of the
// request, whatever its case, or defaultLocale if there is none.
// It is meant for the requests which are not routed, such as the pages not found.
func SaveUnknownLocale(supportedLocales []string, defaultLocale string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := strings.ToLower(c.Request.URL.Path)

		for _, locale := range supportedLocales {
			if strings.Contains(path, "/"+strings.ToLower(locale)+"/") {
				c.Set(localeKey, locale)
				return
			}
		}

		c.Set(localeKey, defaultLocale)
	}
}

// GetLocale returns the locale saved to the context, or an empty string if there is none.
func GetLocale(c *gin.Context) string {
	locale, _ := c.Get(localeKey)
	value, _ := locale.(string)

	return value
}
