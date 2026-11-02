package locale

import (
	"strings"

	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/request"

	"github.com/gin-gonic/gin"
)

// Inject returns a middleware saving the locales to the context: the default and supported locales of the application,
// and the locale of the request, given by the URL parameter or the header, if any.
// It refuses the requests whose locale is not supported.
func Inject(defaultAppLocale string, supportedAppLocales []string) gin.HandlerFunc {
	defaultLocale := strings.ToLower(defaultAppLocale)

	// Create the middleware.
	return func(c *gin.Context) {
		c.Set(defaultLocaleKey, defaultLocale)
		c.Set(supportedLocalesKey, supportedAppLocales)

		// Get the locale defined within the request.
		requestLocale := c.Param(ParameterName)
		if len(requestLocale) == 0 {
			requestLocale = c.Request.Header.Get(request.LocaleHeaderName)
		}
		requestLocale = strings.ToLower(requestLocale)

		if len(requestLocale) > 0 && !IsSupported(c, requestLocale) {
			response.UnsupportedRequestLocale(c, defaultLocale, requestLocale, supportedAppLocales, "error.request.parameter.locale.unsupported.message")
			return
		}

		// Save the locales to the context.
		Set(c, "", requestLocale)
	}
}
