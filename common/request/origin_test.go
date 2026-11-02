package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCrossOriginProtection(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		headers map[string]string
		want    int
	}{
		{"safe method from another site", http.MethodGet, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusOK},
		{"same origin", http.MethodPost, map[string]string{"Sec-Fetch-Site": "same-origin"}, http.StatusOK},
		{"another site", http.MethodPost, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"another origin without Sec-Fetch-Site", http.MethodDelete, map[string]string{"Origin": "https://evil.test"}, http.StatusForbidden},
		{"origin of the host", http.MethodPost, map[string]string{"Origin": "http://web.test"}, http.StatusOK},
		{"trusted origin", http.MethodPost, map[string]string{"Origin": "https://www.example.com", "Sec-Fetch-Site": "cross-site"}, http.StatusOK},
		{"not a browser", http.MethodPost, nil, http.StatusOK},
	}

	middleware, err := CrossOriginProtection("", "https://www.example.com/path")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(middleware)
			engine.Any("/", func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequestWithContext(t.Context(), tt.method, "http://web.test/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestCrossOriginProtection_InvalidOrigin(t *testing.T) {
	for _, origin := range []string{"www.example.com", "://", "https://example.com:x"} {
		if _, err := CrossOriginProtection(origin); err == nil {
			t.Errorf("CrossOriginProtection(%q) must return an error", origin)
		}
	}
}
