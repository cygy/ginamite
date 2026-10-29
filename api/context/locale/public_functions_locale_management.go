package locale

import (
	"slices"

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

// GetLocale : returns the locale bound to the request.
func GetLocale(c *gin.Context) string {
	if value, ok := c.Get(localeKey); ok {
		return value.(string)
	}

	return ""
}

// GetUserLocale : returns the user locale.
func GetUserLocale(c *gin.Context) string {
	if value, ok := c.Get(userLocaleKey); ok {
		return value.(string)
	}

	return ""
}

// GetRequestLocale : returns the context locale.
func GetRequestLocale(c *gin.Context) string {
	if value, ok := c.Get(requestLocaleKey); ok {
		return value.(string)
	}

	return ""
}

func SetLocales(c *gin.Context, userLocale, requestLocale string) {
	c.Set(localeKey, defaultLocale)

	if len(userLocale) > 0 {
		c.Set(userLocaleKey, userLocale)
		c.Set(localeKey, userLocale)
	} else {
		c.Set(userLocaleKey, defaultLocale)
	}

	if len(requestLocale) > 0 {
		c.Set(requestLocaleKey, requestLocale)
		c.Set(localeKey, requestLocale)
	} else {
		c.Set(requestLocaleKey, defaultLocale)
	}
}

func IsLocaleSupported(locale string) bool {
	return slices.Contains(supportedLocales, locale)
}
