package route

import (
	"testing"

	"github.com/cygy/ginamite/web/internal/pagetest"
)

func TestRelativePath(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		locale string
		want   string
		wantOk bool
	}{
		{"one occurrence", "https://web.test/en-us/home", "en-us", "/home", true},
		{"locale only", "https://web.test/en-us", "en-us", "", true},
		{"empty URL", "", "en-us", "", false},
		{"no occurrence", "https://web.test/fr-fr/home", "en-us", "", false},
		{"two occurrences", "https://web.test/en-us/en-us/home", "en-us", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := relativePath(tt.url, tt.locale)

			if got != tt.want || ok != tt.wantOk {
				t.Errorf("relativePath() = %q, %v, want %q, %v", got, ok, tt.want, tt.wantOk)
			}
		})
	}
}

func TestRelativeRoutePath(t *testing.T) {
	pagetest.Setup(t)

	tests := []struct {
		key, locale string
		want        string
		wantPartial bool
		wantOk      bool
	}{
		{"index", "en-us", "/home", false, true},
		{"index", "fr-fr", "/accueil", false, true},
		{"partial", "en-us", "/partial", true, true},
		{"unknown", "en-us", "", true, false},
	}
	for _, tt := range tests {
		path, partial, ok := relativeRoutePath(tt.key, tt.locale)

		if path != tt.want || partial != tt.wantPartial || ok != tt.wantOk {
			t.Errorf("relativeRoutePath(%q, %q) = %q, %v, %v, want %q, %v, %v", tt.key, tt.locale, path, partial, ok, tt.want, tt.wantPartial, tt.wantOk)
		}
	}
}

func TestRelativeAsyncPath(t *testing.T) {
	pagetest.Setup(t)

	tests := []struct {
		key, locale string
		want        string
		wantOk      bool
	}{
		{"like", "en-us", "/async/like/:id", true},
		{"like", "fr-fr", "/async/aimer/:id", true},
		{"list", "fr-fr", "", false},
		{"unknown", "en-us", "", false},
	}
	for _, tt := range tests {
		path, ok := relativeAsyncPath(tt.key, tt.locale)

		if path != tt.want || ok != tt.wantOk {
			t.Errorf("relativeAsyncPath(%q, %q) = %q, %v, want %q, %v", tt.key, tt.locale, path, ok, tt.want, tt.wantOk)
		}
	}
}
