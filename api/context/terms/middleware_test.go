package terms

import (
	"net/http"
	"testing"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

func TestVerifyLatestVersionOfTermsAcceptedByUser(t *testing.T) {
	const secret = "secret"

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	latest := func(version string) func(*gin.Context, string) string {
		return func(*gin.Context, string) string { return version }
	}

	tests := []struct {
		name       string
		authed     bool
		function   func(*gin.Context, string) string
		wantStatus int
	}{
		{"no function is only reported", true, nil, http.StatusOK},
		{"anonymous user is ignored", false, latest("1"), http.StatusOK},
		{"latest version", true, latest("2"), http.StatusOK},
		{"old version is refused", true, latest("1"), http.StatusPreconditionFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ctxLocale.Inject("en", []string{"en"}))
			router.Use(ctxAuthToken.Parse(secret, nil, false))
			router.Use(VerifyLatestVersionOfTermsAcceptedByUser("2", tt.function))
			router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			if tt.authed {
				req.Header.Set("Authorization", "Bearer "+apitest.SignToken(t, secret, "user-1", "token-1", false))
			}

			if got := apitest.Serve(router, req).Code; got != tt.wantStatus {
				t.Errorf("status = %d, want %d", got, tt.wantStatus)
			}
		})
	}
}
