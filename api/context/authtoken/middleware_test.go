package authtoken

import (
	"net/http"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/ctxlocalization"

	tokencache "github.com/cygy/ginamite/api/cache/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

func TestParse(t *testing.T) {
	const secret = "secret"

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})
	revoked := tokencache.NewRevoked()
	revoked.Add(map[string]time.Time{"revoked-token": time.Now().Add(time.Hour)})

	buildUser := func(token *jwt.Token) (any, string, error) { return "the-user", "en", nil }

	tests := []struct {
		name        string
		mustBeThere bool
		header      string
		wantStatus  int
		wantID      string
	}{
		{"missing token is refused when mandatory", true, "", http.StatusUnauthorized, ""},
		{"missing token is accepted when optional", false, "", http.StatusOK, ""},
		{"malformed token", false, "Bearer not-a-token", http.StatusUnauthorized, ""},
		{"wrong secret", false, "Bearer " + apitest.SignToken(t, "other", "u", "t", false), http.StatusUnauthorized, ""},
		{"expired token", false, "Bearer " + apitest.SignToken(t, secret, "u", "t", true), http.StatusUnauthorized, ""},
		{"revoked token", false, "Bearer " + apitest.SignToken(t, secret, "u", "revoked-token", false), http.StatusUnauthorized, "revoked-token"},
		{"valid token", true, "Bearer " + apitest.SignToken(t, secret, "u", "token-1", false), http.StatusOK, "token-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID string
			router := gin.New()
			router.Use(ctxlocalization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}), InjectRevoked(revoked.IsRevoked))
			router.Use(Parse(secret, buildUser, tt.mustBeThere))
			router.GET("/", func(c *gin.Context) {
				gotID = GetID(c)
				c.Status(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			recorder := apitest.Serve(router, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK && gotID != tt.wantID {
				t.Errorf("token ID = %q, want %q", gotID, tt.wantID)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	tests := []struct {
		name       string
		verify     IsValidFunc
		wantStatus int
	}{
		{"no function", nil, http.StatusOK},
		{"valid token", func(*gin.Context, string) (bool, error) { return true, nil }, http.StatusOK},
		{"invalid token", func(*gin.Context, string) (bool, error) { return false, nil }, http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := apitest.NewContext("GET", "/", nil)
			ctxLocale.Set(c, "", "en")
			Verify(tt.verify)(c)
			c.Writer.WriteHeaderNow()

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
