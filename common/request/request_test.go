package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetRealIPAddress(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{"user header first", map[string]string{UserIPAddressHeaderName: "1.1.1.1", IPAddressSourceHeaderName: "2.2.2.2"}, "1.1.1.1"},
		{"proxy header", map[string]string{IPAddressSourceHeaderName: "2.2.2.2"}, "2.2.2.2"},
		{"client address", nil, "192.0.2.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = req

			if got := GetRealIPAddress(c); got != tt.want {
				t.Errorf("GetRealIPAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}
