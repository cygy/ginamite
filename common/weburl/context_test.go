package weburl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInjectAndOf(t *testing.T) {
	gin.SetMode(gin.TestMode)

	conf := load(t)

	tests := []struct {
		name       string
		middleware []gin.HandlerFunc
		want       *Configuration
	}{
		{"injected", []gin.HandlerFunc{Inject(conf)}, conf},
		{"injected nil", []gin.HandlerFunc{Inject(nil)}, nil},
		{"no middleware", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *Configuration
			router := gin.New()
			router.Use(tt.middleware...)
			router.GET("/", func(c *gin.Context) { got = Of(c) })

			router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			if got != tt.want {
				t.Errorf("Of() = %p, want %p", got, tt.want)
			}
		})
	}
}
