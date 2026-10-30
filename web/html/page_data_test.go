package html

import (
	"net/http"
	"testing"
	"time"

	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestPage_InitializeData(t *testing.T) {
	setup(t)

	full := map[string]any{
		"sub": "user1", "nickname": "John", "img1": "1x.png", "img2": "2x.png",
		"unread": float64(3), "timezone": "Asia/Tokyo",
	}
	expired := map[string]any{"sub": "user1", "exp": time.Now().Add(-time.Hour).Unix()}

	tests := []struct {
		name   string
		cookie string
		want   User
		wantTZ string
	}{
		{"all claims", webtest.SignToken(t, testSecret, full), User{ID: "user1", Name: "John", Image1x: "1x.png", Image2x: "2x.png", Unread: 3}, "Asia/Tokyo"},
		{"missing claims", webtest.SignToken(t, testSecret, map[string]any{"sub": "user1"}), User{ID: "user1"}, "Europe/Paris"},
		{"claims of another type", webtest.SignToken(t, testSecret, map[string]any{"sub": 1, "nickname": true, "unread": "3", "timezone": 2}), User{}, "Europe/Paris"},
		{"negative unread", webtest.SignToken(t, testSecret, map[string]any{"sub": "user1", "unread": float64(-4)}), User{ID: "user1"}, "Europe/Paris"},
		{"wrong secret", webtest.SignToken(t, "other", full), User{}, "Europe/Paris"},
		{"expired", webtest.SignToken(t, testSecret, expired), User{}, "Europe/Paris"},
		{"not a token", "garbage", User{}, "Europe/Paris"},
		{"no cookie", "", User{}, "Europe/Paris"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := requestFor("en-us")
			if tt.cookie != "" {
				c.Request.AddCookie(&http.Cookie{Name: testCookieName, Value: tt.cookie})
			}

			page := Main.preprocessedPages["home"]["en-us"].Copy()
			page.Vars = map[string]any{"a": 1}
			page.Content.Error.Message = "stale"
			page.InitializeData(c)

			if page.Content.CurrentUser != tt.want {
				t.Errorf("CurrentUser = %+v, want %+v", page.Content.CurrentUser, tt.want)
			}
			if page.Timezone != tt.wantTZ {
				t.Errorf("Timezone = %q, want %q", page.Timezone, tt.wantTZ)
			}
			if page.Content.Title != "Home" || page.Var("a") != 1 || page.Content.HasError() {
				t.Errorf("Content = %+v", page.Content)
			}
		})
	}
}
