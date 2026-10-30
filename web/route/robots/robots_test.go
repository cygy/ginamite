package robots

import (
	"net/http"
	"net/http/httptest"
	"testing"

	r "github.com/cygy/ginamite/web/robots"

	"github.com/gin-gonic/gin"
)

func TestRobots(t *testing.T) {
	old := r.Content
	r.Content = "User-agent: *\nDisallow:\n"
	t.Cleanup(func() { r.Content = old })

	engine := gin.New()
	engine.GET("/robots.txt", Robots)

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	if rec.Code != http.StatusOK || rec.Body.String() != r.Content {
		t.Errorf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
}
