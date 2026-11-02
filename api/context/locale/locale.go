// Package locale stores and reads the user and request locales in the gin context.
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

// Get returns the locale bound to the request.
func Get(c *gin.Context) string {
	if value, ok := c.Get(localeKey); ok {
		return value.(string)
	}

	return ""
}

// GetUser returns the user locale.
func GetUser(c *gin.Context) string {
	if value, ok := c.Get(userLocaleKey); ok {
		return value.(string)
	}

	return ""
}

// GetRequest returns the context locale.
func GetRequest(c *gin.Context) string {
	if value, ok := c.Get(requestLocaleKey); ok {
		return value.(string)
	}

	return ""
}

// Set stores the user and request locales in the context.
// The default locale of the context, see Inject, is used when a locale is empty.
func Set(c *gin.Context, userLocale, requestLocale string) {
	defaultLocale := defaultOf(c)
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

// defaultOf returns the default locale saved in the context by Inject, or an empty string if there is none.
func defaultOf(c *gin.Context) string {
	if value, ok := c.Get(defaultLocaleKey); ok {
		return value.(string)
	}

	return ""
}

// IsSupported returns true if the locale is one of the supported locales of the context, see Inject.
func IsSupported(c *gin.Context, locale string) bool {
	if value, ok := c.Get(supportedLocalesKey); ok {
		return slices.Contains(value.([]string), locale)
	}

	return false
}
