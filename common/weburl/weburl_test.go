package weburl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

const routes = `
routes:
  account.profile:
    paths:
      fr-fr: /{{.Locale}}/compte
      en-gb: /{{.Locale}}/account
      de: /konto
  user.show:
    paths:
      fr-fr: /fr/utilisateur/:id
`

func writeRoutes(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "routes.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func load(t *testing.T) *Configuration {
	t.Helper()

	conf := NewConfiguration()
	if err := conf.Initialize(writeRoutes(t, routes), "https://example.com"); err != nil {
		t.Fatal(err)
	}

	return conf
}

func TestNewConfiguration(t *testing.T) {
	if conf := NewConfiguration(); conf.Routes == nil || len(conf.Routes) != 0 {
		t.Errorf("NewConfiguration() = %+v", conf)
	}
}

func TestConfiguration_Initialize(t *testing.T) {
	conf := load(t)

	profile := conf.Routes["account.profile"].AbsolutePaths
	want := map[string]string{
		"fr-fr": "https://example.com/fr-fr/compte",
		"fr":    "https://example.com/fr-fr/compte",
		"en-gb": "https://example.com/en-gb/account",
		"en":    "https://example.com/en-gb/account",
		"de":    "https://example.com/konto",
	}
	for locale, url := range want {
		if profile[locale] != url {
			t.Errorf("AbsolutePaths[%q] = %q, want %q", locale, profile[locale], url)
		}
	}
}

func TestConfiguration_Initialize_Errors(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"missing file", filepath.Join(t.TempDir(), "missing.yml")},
		{"invalid yaml", writeRoutes(t, "routes: [")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewConfiguration().Initialize(tt.file, "https://example.com"); err == nil {
				t.Error("Initialize() must return an error")
			}
			if conf, err := Load(tt.file, "https://example.com"); err == nil || conf != nil {
				t.Errorf("Load() = %v, %v, want an error", conf, err)
			}
		})
	}
}

func TestConfiguration_GetURL(t *testing.T) {
	conf := load(t)

	tests := []struct {
		name, key, locale, want string
		wantWarning             bool
	}{
		{"known", "account.profile", "fr-fr", "https://example.com/fr-fr/compte", false},
		{"language only", "account.profile", "en", "https://example.com/en-gb/account", false},
		{"unknown locale", "account.profile", "es", "", true},
		{"unknown route", "nope", "fr-fr", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conf.GetURL(tt.key, tt.locale); got != tt.want {
				t.Errorf("GetURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoute_HasVariable(t *testing.T) {
	conf := load(t)

	profile := conf.Routes["account.profile"]
	if profile.HasVariable() {
		t.Error("account.profile has no variable")
	}

	route := conf.Routes["user.show"]
	if !route.HasVariable() {
		t.Error("user.show has a variable")
	}

	if (&Route{}).HasVariable() {
		t.Error("a route without path has no variable")
	}
}

func TestLoad(t *testing.T) {
	conf, err := Load(writeRoutes(t, routes), "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	if got := conf.GetURL("account.profile", "de"); got != "https://example.com/konto" {
		t.Errorf("GetURL() = %q", got)
	}
}

func TestConfiguration_GetURL_NilConfiguration(t *testing.T) {
	var conf *Configuration

	if got := conf.GetURL("account.profile", "fr-fr"); got != "" {
		t.Errorf("GetURL() = %q, want an empty string", got)
	}
}

func TestConfiguration_GetURLWithParams(t *testing.T) {
	conf := NewConfiguration()
	conf.Routes["order"] = Route{AbsolutePaths: map[string]string{"en": "https://example.com/en/order/:id/:slug/:identifier"}}
	var nilConf *Configuration

	tests := []struct {
		name   string
		conf   *Configuration
		key    string
		params map[string]string
		want   string
	}{
		{"all the parameters", conf, "order", map[string]string{"id": "1", "slug": "a-b", "identifier": "x"}, "https://example.com/en/order/1/a-b/x"},
		{"with colon", conf, "order", map[string]string{":id": "1", ":slug": "s", ":identifier": "i"}, "https://example.com/en/order/1/s/i"},
		{"missing value left as is", conf, "order", map[string]string{"id": "1"}, "https://example.com/en/order/1/:slug/:identifier"},
		{"unknown parameter ignored", conf, "order", map[string]string{"other": "1"}, "https://example.com/en/order/:id/:slug/:identifier"},
		{"no parameters", conf, "order", nil, "https://example.com/en/order/:id/:slug/:identifier"},
		{"unknown route", conf, "missing", map[string]string{"id": "1"}, ""},
		{"nil configuration", nilConf, "order", map[string]string{"id": "1"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conf.GetURLWithParams(tt.key, "en", tt.params); got != tt.want {
				t.Errorf("GetURLWithParams() = %q, want %q", got, tt.want)
			}
		})
	}
}
