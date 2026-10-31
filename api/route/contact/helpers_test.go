package contact

import (
	"net/http"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/ctxlocalization"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

// testFunctions are the functions of the handlers, reset at the end of each test.
var testFunctions Functions

func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})
	t.Cleanup(func() { testFunctions = Functions{} })
}

func serve(handler gin.HandlerFunc, method, target, body string, middlewares ...gin.HandlerFunc) int {
	router := gin.New()
	router.Use(Inject(&testFunctions), ctxlocalization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}))
	router.Use(middlewares...)
	router.Any("/:id", handler)
	router.Any("/", handler)

	req, _ := http.NewRequest(method, target, strings.NewReader(body))

	return apitest.Serve(router, req).Code
}
