package robots

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRobots(t *testing.T) {
	const content = "User-agent: *\nDisallow:\n"

	engine := gin.New()
	engine.GET("/robots.txt", Robots(content))

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	if rec.Code != http.StatusOK || rec.Body.String() != content {
		t.Errorf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
}
