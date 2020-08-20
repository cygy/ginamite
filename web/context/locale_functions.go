package context

import (
	"strings"

	"github.com/gin-gonic/gin"
)

/*
	Here are defined the locales used within the request.

	The 'user locale' is the locale defined by a registered user.
		Once defined, this locale overrides the default app locale.
	The 'request locale' is the locale defined within the request (header or parameter).
		It is used to define the locale used for the output (e.g. "I want this resource in this locale.")
	The 'locale' is defined by order: the 'request locale', the 'user locale', the 'default locale'.
*/

// SaveLocale : saves the locale of the target page to the context.
func SaveLocale(contextLocale string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(localeKey, contextLocale)
	}
}

// SaveUnknownLocale : saves the locale of the target page to the context.
func SaveUnknownLocale(supportedLocales []string, defaultLocale string) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := strings.ToLower(c.Request.URL.Path)

		for _, locale := range supportedLocales {
			if strings.Contains(url, "/"+locale+"/") {
				c.Set(localeKey, locale)
				return
			}
		}

		c.Set(localeKey, defaultLocale)
	}
}

// GetLocale : returns the locale from the context.
func GetLocale(c *gin.Context) string {
	if value, ok := c.Get(localeKey); ok {
		return value.(string)
	}

	return ""
}
