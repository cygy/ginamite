package context

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

const (
	testCookie = "auth"
	testSecret = "secret"
)

// serveWithToken runs SaveAuthToken then the handler on a request carrying the cookie value, if not empty.
func serveWithToken(cookie string, handler gin.HandlerFunc) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: testCookie, Value: cookie})
	}

	webtest.Serve(req, SaveAuthToken(testCookie, testSecret), handler)
}

func TestSaveAuthToken(t *testing.T) {
	claims := map[string]any{"sub": "user1", "id": "token1", "up": float64(1700000000)}
	tests := []struct {
		name    string
		cookie  string
		wantID  string
		wantLog bool
	}{
		{"valid", webtest.SignToken(t, testSecret, claims), "token1", false},
		{"no cookie", "", "", false},
		{"wrong secret", webtest.SignToken(t, "other", claims), "", true},
		{"not a token", "garbage", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webtest.Hook.Reset()

			var id, userID string
			var authenticated bool
			var lastUpdate time.Time
			serveWithToken(tt.cookie, func(c *gin.Context) {
				id = GetAuthTokenID(c)
				userID = GetUserID(c)
				authenticated = IsUserAuthenticated(c)
				lastUpdate = GetAuthTokenLastUpdate(c)
			})

			if id != tt.wantID {
				t.Errorf("GetAuthTokenID() = %q, want %q", id, tt.wantID)
			}
			if wantAuthenticated := tt.wantID != ""; authenticated != wantAuthenticated {
				t.Errorf("IsUserAuthenticated() = %v, want %v", authenticated, wantAuthenticated)
			}
			if tt.wantID != "" {
				if userID != "user1" {
					t.Errorf("GetUserID() = %q, want user1", userID)
				}
				if lastUpdate.Unix() != 1700000000 {
					t.Errorf("GetAuthTokenLastUpdate() = %v", lastUpdate)
				}
			} else if userID != "" || lastUpdate.Unix() != 0 {
				t.Errorf("GetUserID() = %q, GetAuthTokenLastUpdate() = %v, want none", userID, lastUpdate)
			}
			if logged := len(webtest.Hook.AllEntries()) > 0; logged != tt.wantLog {
				t.Errorf("logged = %v, want %v", logged, tt.wantLog)
			}
		})
	}
}

func TestSaveAuthToken_DoesNotLogTheToken(t *testing.T) {
	webtest.Hook.Reset()

	token := webtest.SignToken(t, "other", map[string]any{"sub": "user1"})
	serveWithToken(token, func(c *gin.Context) {})

	for _, entry := range webtest.Hook.AllEntries() {
		if strings.Contains(entry.Message, token) {
			t.Errorf("the message contains the token: %q", entry.Message)
		}
		for name, value := range entry.Data {
			if s, ok := value.(string); ok && strings.Contains(s, token) {
				t.Errorf("the field %q contains the token", name)
			}
		}
	}
}

func TestGetJWTToken_IgnoresOtherTypes(t *testing.T) {
	c, _ := webtest.Context(http.MethodGet, "/", nil)
	c.Set(tokenKey, "not a token")

	if token := getJWTToken(c); token != nil {
		t.Errorf("getJWTToken() = %v, want nil", token)
	}
	if GetUserID(c) != "" || IsUserAuthenticated(c) {
		t.Error("a value which is not a token must not authenticate")
	}
}
