package locale

import (
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"
)

func TestSetAndGet(t *testing.T) {
	defaultLocale = "en"

	tests := []struct {
		name                                 string
		userLocale, requestLocale            string
		wantLocale, wantUser, wantRequestLoc string
	}{
		{"defaults", "", "", "en", "en", "en"},
		{"user locale overrides the default", "fr", "", "fr", "fr", "en"},
		{"request locale overrides the user locale", "fr", "es", "es", "fr", "es"},
		{"request locale only", "", "es", "es", "en", "es"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := apitest.NewContext("GET", "/", nil)
			Set(c, tt.userLocale, tt.requestLocale)

			if got := Get(c); got != tt.wantLocale {
				t.Errorf("Get() = %q, want %q", got, tt.wantLocale)
			}
			if got := GetUser(c); got != tt.wantUser {
				t.Errorf("GetUser() = %q, want %q", got, tt.wantUser)
			}
			if got := GetRequest(c); got != tt.wantRequestLoc {
				t.Errorf("GetRequest() = %q, want %q", got, tt.wantRequestLoc)
			}
		})
	}
}

func TestGetters_NothingStored(t *testing.T) {
	c, _ := apitest.NewContext("GET", "/", nil)

	if Get(c) != "" || GetUser(c) != "" || GetRequest(c) != "" {
		t.Error("getters must return an empty string when no locale is stored")
	}
}

func TestIsSupported(t *testing.T) {
	supportedLocales = []string{"en", "fr"}

	tests := []struct {
		locale string
		want   bool
	}{
		{"en", true},
		{"fr", true},
		{"es", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			if got := IsSupported(tt.locale); got != tt.want {
				t.Errorf("IsSupported(%q) = %v, want %v", tt.locale, got, tt.want)
			}
		})
	}
}
