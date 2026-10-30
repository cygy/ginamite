package request

import (
	"testing"

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

// ctxLocaleMiddleware stores the "en" locale in the context.
func ctxLocaleMiddleware() gin.HandlerFunc {
	return ctxLocale.Inject("en", []string{"en", "fr"})
}
