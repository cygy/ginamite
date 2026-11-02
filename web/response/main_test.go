package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/internal/pagetest"

	"github.com/gin-gonic/gin"
)

// serve initializes the pages then sends a GET request to the handler with the locale and the headers, and returns
// the response.
func serve(t *testing.T, locale string, headers map[string]string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	engine := pagetest.Setup(t)
	engine.GET("/test", context.SaveLocale(locale), handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	return rec
}
