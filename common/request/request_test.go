package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestConfigureClientIP(t *testing.T) {
	tests := []struct {
		name    string
		proxies []string
		headers map[string]string
		want    string
	}{
		{"no proxy trusted: the headers are ignored", nil, map[string]string{UserIPAddressHeaderName: "1.1.1.1", "X-Forwarded-For": "2.2.2.2", "X-Real-IP": "3.3.3.3"}, "10.1.2.3"},
		{"untrusted proxy: the headers are ignored", []string{"192.168.0.0/16"}, map[string]string{UserIPAddressHeaderName: "1.1.1.1"}, "10.1.2.3"},
		{"trusted proxy: the user header first", []string{"10.0.0.0/8"}, map[string]string{UserIPAddressHeaderName: "1.1.1.1", "X-Forwarded-For": "2.2.2.2"}, "1.1.1.1"},
		{"trusted proxy: X-Forwarded-For", []string{"10.0.0.0/8"}, map[string]string{"X-Forwarded-For": "2.2.2.2"}, "2.2.2.2"},
		{"trusted proxy: X-Real-IP", []string{"10.0.0.0/8"}, map[string]string{"X-Real-IP": "3.3.3.3"}, "3.3.3.3"},
		{"trusted proxy: invalid header", []string{"10.0.0.0/8"}, map[string]string{UserIPAddressHeaderName: "nope"}, "10.1.2.3"},
		{"trusted proxy without header", []string{"10.0.0.0/8"}, nil, "10.1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := gin.New()
			if err := ConfigureClientIP(engine, tt.proxies); err != nil {
				t.Fatal(err)
			}

			var got string
			engine.GET("/", func(c *gin.Context) { got = c.ClientIP() })

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			req.RemoteAddr = "10.1.2.3:1234"
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			engine.ServeHTTP(httptest.NewRecorder(), req)

			if got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfigureClientIP_Invalid(t *testing.T) {
	if err := ConfigureClientIP(gin.New(), []string{"not a network"}); err == nil {
		t.Error("an invalid proxy must be rejected")
	}
}
