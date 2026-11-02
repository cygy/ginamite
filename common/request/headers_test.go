package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name     string
		hsts     bool
		wantHSTS string
	}{
		{"without HSTS", false, ""},
		{"with HSTS", true, hstsValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(SecurityHeaders(tt.hsts))
			engine.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

			want := map[string]string{
				"X-Content-Type-Options":    "nosniff",
				"X-Frame-Options":           "DENY",
				"Content-Security-Policy":   "frame-ancestors 'none'",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
				"Strict-Transport-Security": tt.wantHSTS,
			}
			for name, value := range want {
				if got := rec.Header().Get(name); got != value {
					t.Errorf("%s = %q, want %q", name, got, value)
				}
			}
		})
	}
}

func TestSecurityHeaders_Override(t *testing.T) {
	engine := gin.New()
	engine.Use(SecurityHeaders(false))
	engine.GET("/", func(c *gin.Context) {
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Errorf("X-Frame-Options = %q, want the value of the handler", got)
	}
}

func TestSecurityHeaders_AbortedRequest(t *testing.T) {
	engine := gin.New()
	engine.Use(SecurityHeaders(false), func(c *gin.Context) { c.AbortWithStatus(http.StatusForbidden) })
	engine.GET("/", func(*gin.Context) {})

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q on an aborted request, want nosniff", got)
	}
}
