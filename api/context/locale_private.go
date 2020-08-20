package context

import "github.com/gin-gonic/gin"

func setLocales(c *gin.Context, userLocale, requestLocale string) {
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

func isLocaleSupported(locale string) bool {
	found := false
	for _, supportedLocale := range supportedLocales {
		if supportedLocale == locale {
			found = true
			break
		}
	}

	return found
}
