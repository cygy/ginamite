package authtoken

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/ctxkey"

	tokencache "github.com/cygy/ginamite/api/cache/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/log"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	logtest "github.com/sirupsen/logrus/hooks/test"
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
			router.Use(ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}), InjectRevoked(revoked.IsRevoked))
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
		{"valid token", func(context.Context, string) (bool, error) { return true, nil }, http.StatusOK},
		{"invalid token", func(context.Context, string) (bool, error) { return false, nil }, http.StatusUnauthorized},
		{"token which can not be verified", func(context.Context, string) (bool, error) { return false, errors.New("boom") }, http.StatusInternalServerError},
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

func TestParse_ExpiredTokenAndLogs(t *testing.T) {
	const secret = "secret"

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	hook := logtest.NewLocal(log.Main.Logger)
	defer hook.Reset()

	router := gin.New()
	router.Use(ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}), InjectRevoked(func(string) bool { return false }))
	router.Use(Parse(secret, nil, true))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	serve := func(token string) string {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		return apitest.Serve(router, req).Body.String()
	}

	expired := apitest.SignToken(t, secret, "u", "t", true)
	invalid := apitest.SignToken(t, "other", "u", "t", false)

	expiredBody, invalidBody := serve(expired), serve(invalid)
	if want := fmt.Sprintf(`"code":%d`, auth.ExpiredAuthorizationToken); !strings.Contains(expiredBody, want) {
		t.Errorf("expired token: body = %s, want %s", expiredBody, want)
	}
	if invalidBody == expiredBody {
		t.Errorf("an invalid token is reported as an expired one: %s", invalidBody)
	}

	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, expired) || strings.Contains(entry.Message, invalid) {
			t.Errorf("a token is logged: %q", entry.Message)
		}
		for _, value := range entry.Data {
			if text, ok := value.(string); ok && (strings.Contains(text, expired) || strings.Contains(text, invalid)) {
				t.Errorf("a token is logged: %v", entry.Data)
			}
		}
	}
}
