package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite/web/route"

	"github.com/gin-gonic/gin"
)

func TestConfigureRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		exists bool
		want   int
	}{
		{"route defined by the routes file", true, http.StatusInternalServerError},
		{"route missing from the routes file", false, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(gin.CustomRecovery(func(c *gin.Context, _ any) { c.AbortWithStatus(http.StatusInternalServerError) }))
			configureRoutes(engine.Group("/en-us"), "en-us", route.Middlewares{}, func(key string) (string, bool, gin.HandlerFunc, bool) {
				if key != "index" {
					t.Errorf("key = %q, want index", key)
				}

				return "/hello", false, func(*gin.Context) {}, tt.exists
			})

			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/en-us/hello", nil))

			// The handler panics without the pages of the server in the context: it is registered.
			if recorder.Code != tt.want {
				t.Errorf("status = %d, want %d", recorder.Code, tt.want)
			}
		})
	}
}
