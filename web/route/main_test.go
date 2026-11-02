package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/internal/pagetest"

	"github.com/gin-gonic/gin"
)

// testClient is the client of the API used by the routers, set by the tests.
var testClient *api.Client

// newRouter initializes the pages and returns a router of their engine, with the default routes not loaded.
func newRouter(t *testing.T) (*Router, *gin.Engine) {
	t.Helper()

	engine := pagetest.Setup(t)

	t.Cleanup(func() { testClient = nil })

	router := NewRouter(engine, pagetest.Pages, testClient, pagetest.Locales, "en-us", pagetest.CookieName, pagetest.Secret)
	router.Robots = "User-agent: *"

	return router, engine
}

// do sends the request to the engine and returns the response.
func do(engine *gin.Engine, method, target string, mutate ...func(*http.Request)) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for _, m := range mutate {
		m(req)
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	return rec
}
