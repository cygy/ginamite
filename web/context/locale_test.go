package context

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestSaveLocale(t *testing.T) {
	c, _ := webtest.Context(http.MethodGet, "/", nil)

	if got := GetLocale(c); got != "" {
		t.Errorf("GetLocale() = %q before SaveLocale, want empty", got)
	}

	SaveLocale("fr-fr")(c)

	if got := GetLocale(c); got != "fr-fr" {
		t.Errorf("GetLocale() = %q, want fr-fr", got)
	}
}

func TestGetLocale_IgnoresOtherTypes(t *testing.T) {
	c, _ := webtest.Context(http.MethodGet, "/", nil)
	c.Set(localeKey, 42)

	if got := GetLocale(c); got != "" {
		t.Errorf("GetLocale() = %q, want empty", got)
	}
}

func TestSaveUnknownLocale(t *testing.T) {
	locales := []string{"en-us", "fr-FR"}
	tests := []struct {
		path string
		want string
	}{
		{"/en-us/missing", "en-us"},
		{"/fr-fr/some/page", "fr-FR"},
		{"/FR-FR/Page", "fr-FR"},
		{"/de-de/page", "en-us"},
		{"/", "en-us"},
		{"/page/en-us/", "en-us"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			c, _ := webtest.Context(http.MethodGet, tt.path, nil)

			SaveUnknownLocale(locales, "en-us")(c)

			if got := GetLocale(c); got != tt.want {
				t.Errorf("GetLocale() = %q, want %q", got, tt.want)
			}
		})
	}
}
