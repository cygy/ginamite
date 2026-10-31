package locale

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/ctxlocalization"

	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/request"

	"github.com/gin-gonic/gin"
)

func TestInject(t *testing.T) {
	apitest.InitLocalization(t, map[string]string{
		"error.request.parameter.locale.unsupported.message":  "unsupported",
		"error.request.parameter.locale.unsupported.reason":   "unsupported {{.Locale}}",
		"error.request.parameter.locale.unsupported.recovery": "use {{.Locales}}",
	})

	tests := []struct {
		name       string
		target     string
		header     string
		wantStatus int
		wantLocale string
	}{
		{"no locale uses the default", "/", "", http.StatusOK, "en"},
		{"locale from the header", "/", "FR", http.StatusOK, "fr"},
		{"locale from the URL parameter", "/fr", "", http.StatusOK, "fr"},
		{"unsupported locale", "/", "es", http.StatusBadRequest, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			router := gin.New()
			router.Use(ctxlocalization.Inject(apitest.Localization), Inject("EN", []string{"en", "fr"}))
			handler := func(c *gin.Context) {
				got = Get(c)
				c.Status(http.StatusOK)
			}
			router.GET("/", handler)
			router.GET("/:"+ParameterName, handler)

			req, _ := http.NewRequest(http.MethodGet, tt.target, nil)
			if tt.header != "" {
				req.Header.Set(request.LocaleHeaderName, tt.header)
			}
			recorder := apitest.Serve(router, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if got != tt.wantLocale {
				t.Errorf("locale = %q, want %q", got, tt.wantLocale)
			}
		})
	}
}
