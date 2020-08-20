package context

import (
	"strings"

	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/request"

	"github.com/gin-gonic/gin"
)

// InjectLocale saves the current locales to the context.
func InjectLocale(defaultAppLocale string, supportedAppLocales []string) gin.HandlerFunc {
	// First the app locales are defined.
	defaultLocale = strings.ToLower(defaultAppLocale)
	supportedLocales = supportedAppLocales

	// Create the middleware.
	return func(c *gin.Context) {
		// Get the locale defined within the request.
		requestLocale := c.Param(LocaleParameterName)

		if len(requestLocale) == 0 {
			requestLocale = c.Request.Header.Get(request.LocaleHeaderName)
		}

		requestLocale = strings.ToLower(requestLocale)

		if len(requestLocale) > 0 && !isLocaleSupported(requestLocale) {
			response.UnsupportedRequestLocale(c, defaultLocale, requestLocale, supportedAppLocales, "error.request.parameter.locale.unsupported.message")
			return
		}

		// Save the locales to the context.
		setLocales(c, "", requestLocale)
	}
}
