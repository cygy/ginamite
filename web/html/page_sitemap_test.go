package html

import "testing"

func TestSitemapUpdateFrequencyString(t *testing.T) {
	if got := SitemapUpdateFrequency(SitemapUpdateFrequencyWeekly).String(); got != "weekly" {
		t.Errorf("String() = %q, want weekly", got)
	}
}

func TestSitemapIsDefined(t *testing.T) {
	tests := []struct {
		name    string
		sitemap Sitemap
		want    bool
	}{
		{"empty", Sitemap{}, false},
		{"included only", Sitemap{Included: true, Priority: 1}, false},
		{"with an update frequency", Sitemap{UpdateFrequency: SitemapUpdateFrequencyDaily}, true},
	}
	for _, tt := range tests {
		if got := tt.sitemap.IsDefined(); got != tt.want {
			t.Errorf("%s: IsDefined() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
