package request

import (
	"testing"

	"github.com/cygy/ginamite/common/ctxlocalization"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

func initLocalization(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{
		"msg": "message", "reason": "reason", "recovery": "recovery",
		"error.request.parameter.offset.negative.message":  "m",
		"error.request.parameter.offset.negative.reason":   "r",
		"error.request.parameter.offset.negative.recovery": "r",
		"error.request.parameter.limit.negative.message":   "m",
		"error.request.parameter.limit.negative.reason":    "r",
		"error.request.parameter.limit.negative.recovery":  "r",
		"error.request.parameter.limit.max.message":        "m",
		"error.request.parameter.limit.max.reason":         "max {{.LimitMax}}",
		"error.request.parameter.limit.max.recovery":       "max {{.LimitMax}}",
	})
}

// ctxLocaleMiddleware stores the "en" locale and the translation engine in the context.
func ctxLocaleMiddleware() gin.HandlerFunc {
	locale := ctxLocale.Inject("en", []string{"en", "fr"})

	return func(c *gin.Context) {
		ctxlocalization.Set(c, apitest.Localization)
		locale(c)
	}
}
