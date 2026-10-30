package contact

import (
	"net/http"
	"strings"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})
	t.Cleanup(func() {
		CreateContact, UpdateContact, GetContacts, GetContact = nil, nil, nil, nil
	})
}

func serve(handler gin.HandlerFunc, method, target, body string, middlewares ...gin.HandlerFunc) int {
	router := gin.New()
	router.Use(ctxLocale.Inject("en", []string{"en"}))
	router.Use(middlewares...)
	router.Any("/:id", handler)
	router.Any("/", handler)

	req, _ := http.NewRequest(method, target, strings.NewReader(body))

	return apitest.Serve(router, req).Code
}
