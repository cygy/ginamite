package config

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"
)

func TestNewConfiguration(t *testing.T) {
	c := NewConfiguration()

	if c.Environment != environment.Development || c.Version != version.Beta {
		t.Errorf("environment/version = %s/%s", c.Environment, c.Version)
	}
	if c.Server.Host != "localhost" || c.Server.Port != 80 || c.Server.MaxBodySize != DefaultMaxBodySize || c.DefaultLocale != "en-us" || c.JWT.SigningMethod != "HS256" {
		t.Errorf("unexpected defaults %+v", c)
	}
	if c.Locales == nil || c.SupportedLocales == nil || c.FallbackLocales == nil {
		t.Error("the collections must be initialized")
	}
}

func TestConfiguration_Initialize(t *testing.T) {
	file := writeFile(t, `
default_locale: fr-fr
api_features: [auth, contact]
locales:
  fr-fr:
    default: true
    date_format: DD/MM/YYYY
  en-gb:
    default: true
  en-us: {}
  de: {default: true}
server:
  port: 8080
broker:
  service: kafka
  host: kafka
  port: 9092
worker_consumers:
  - topic: mail
    group: mails
  - topic: task
`)

	c := NewConfiguration()
	if err := c.Initialize(file, environment.Production, version.Final); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if c.Environment != environment.Production || c.Version != version.Final {
		t.Errorf("environment/version = %s/%s", c.Environment, c.Version)
	}
	if !slices.Equal(c.APIFeatures, []string{"auth", "contact"}) {
		t.Errorf("API features = %v", c.APIFeatures)
	}
	if c.Server.Port != 8080 || c.Server.Host != "localhost" {
		t.Errorf("server = %+v, want the file value merged over the defaults", c.Server)
	}
	if c.Locales["fr-fr"].DateFormat != "DD/MM/YYYY" || len(c.SupportedLocales) != 4 {
		t.Errorf("locales = %+v, supported = %v", c.Locales, c.SupportedLocales)
	}
	if c.FallbackLocales["fr"] != "fr-fr" || c.FallbackLocales["en"] != "en-gb" {
		t.Errorf("fallback locales = %v", c.FallbackLocales)
	}
	if _, ok := c.FallbackLocales["de"]; ok {
		t.Error("a locale without region must not be a fallback")
	}
	if want := (BrokerConfiguration{Service: "kafka", Host: "kafka", Port: 9092}); c.Broker != want {
		t.Errorf("broker = %+v, want %+v", c.Broker, want)
	}
	if want := []ConsumerConfiguration{{Topic: "mail", Group: "mails"}, {Topic: "task"}}; !slices.Equal(c.WorkerConsumers, want) {
		t.Errorf("worker consumers = %+v, want %+v", c.WorkerConsumers, want)
	}
}

func TestConfiguration_Initialize_Errors(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"missing file", filepath.Join(t.TempDir(), "missing.yml")},
		{"invalid yaml", writeFile(t, "server: [")},
		{"unknown same site", writeFile(t, "cookie:\n  same_site: maybe")},
		{"same site none without secure", writeFile(t, "cookie:\n  same_site: none\n  secure: false")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewConfiguration().Initialize(tt.file, environment.Test, version.Final); err == nil {
				t.Error("Initialize() must return an error")
			}
		})
	}
}

func TestConfiguration_SupportsLocale(t *testing.T) {
	c := &Configuration{SupportedLocales: []string{"fr-fr", "en-us"}}

	tests := []struct {
		locale string
		want   bool
	}{
		{"fr-fr", true},
		{"FR-FR", true},
		{"en-us", true},
		{"de-de", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			if got := c.SupportsLocale(tt.locale); got != tt.want {
				t.Errorf("SupportsLocale(%q) = %v, want %v", tt.locale, got, tt.want)
			}
		})
	}
}

func TestConfiguration_Predicates(t *testing.T) {
	tests := []struct {
		name string
		c    Configuration
		got  func(*Configuration) bool
		want bool
	}{
		{"test environment", Configuration{Environment: environment.Test}, (*Configuration).IsTestEnvironment, true},
		{"not test environment", Configuration{Environment: environment.Production}, (*Configuration).IsTestEnvironment, false},
		{"development environment", Configuration{Environment: environment.Development}, (*Configuration).IsDevelopmentEnvironment, true},
		{"not development environment", Configuration{}, (*Configuration).IsDevelopmentEnvironment, false},
		{"production environment", Configuration{Environment: environment.Production}, (*Configuration).IsProductionEnvironment, true},
		{"not production environment", Configuration{Environment: environment.Test}, (*Configuration).IsProductionEnvironment, false},
		{"debug in development", Configuration{Environment: environment.Development}, (*Configuration).IsDebugModeEnabled, true},
		{"no debug in production", Configuration{Environment: environment.Production}, (*Configuration).IsDebugModeEnabled, false},
		{"alpha", Configuration{Version: version.Alpha}, (*Configuration).IsAlphaVersion, true},
		{"not alpha", Configuration{Version: version.Beta}, (*Configuration).IsAlphaVersion, false},
		{"beta", Configuration{Version: version.Beta}, (*Configuration).IsBetaVersion, true},
		{"not beta", Configuration{Version: version.Final}, (*Configuration).IsBetaVersion, false},
		{"final", Configuration{Version: version.Final}, (*Configuration).IsFinalVersion, true},
		{"not final", Configuration{Version: version.Alpha}, (*Configuration).IsFinalVersion, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.got(&tt.c); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// writeFile writes content in a temporary file and returns its path.
func writeFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestNewConfiguration_RateLimits(t *testing.T) {
	c := NewConfiguration()

	if c.APIRateLimiter.Limits.Authentication.Requests != 10 || c.APIRateLimiter.Limits.Authentication.Seconds != 60 {
		t.Errorf("authentication = %+v", c.APIRateLimiter.Limits.Authentication)
	}
	if c.APIRateLimiter.Limits.Registration.Requests == 0 || c.APIRateLimiter.Limits.ForgotPassword.Requests == 0 {
		t.Errorf("limits = %+v", c.APIRateLimiter.Limits)
	}
	if c.Server.TrustedProxies == nil || len(c.Server.TrustedProxies) != 0 {
		t.Errorf("trusted proxies = %v, want none", c.Server.TrustedProxies)
	}
}

func TestConfiguration_Section(t *testing.T) {
	file := writeFile(t, `
myplugin:
  retries: 5
  name: !env MYPLUGIN_NAME
`)
	t.Setenv("MYPLUGIN_NAME", "from env")

	c := NewConfiguration()
	if err := c.Initialize(file, environment.Test, version.Final); err != nil {
		t.Fatal(err)
	}

	t.Run("present", func(t *testing.T) {
		settings := struct {
			Retries int
			Name    string
			Other   string
		}{Retries: 1, Other: "default"}

		if err := c.Section("myplugin", &settings); err != nil {
			t.Fatal(err)
		}
		if settings.Retries != 5 || settings.Name != "from env" || settings.Other != "default" {
			t.Errorf("settings = %+v, want the file merged over the defaults", settings)
		}
	})

	t.Run("absent", func(t *testing.T) {
		settings := struct{ Retries int }{Retries: 1}

		if err := c.Section("missing", &settings); err != nil || settings.Retries != 1 {
			t.Errorf("Section() = %v, settings = %+v, want the defaults", err, settings)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		settings := struct{ Retries []string }{}

		if err := c.Section("myplugin", &settings); err == nil || !strings.Contains(err.Error(), `"myplugin"`) {
			t.Errorf("Section() error = %v", err)
		}
	})

	t.Run("no file", func(t *testing.T) {
		if err := NewConfiguration().Section("myplugin", &struct{}{}); err != nil {
			t.Errorf("Section() error = %v", err)
		}
	})
}

func TestCookieConfiguration_SameSite(t *testing.T) {
	tests := []struct {
		setting string
		secure  bool
		want    http.SameSite
		wantErr bool
	}{
		{"lax", false, http.SameSiteLaxMode, false},
		{"Strict", false, http.SameSiteStrictMode, false},
		{"none", true, http.SameSiteNoneMode, false},
		{"none", false, http.SameSiteNoneMode, true},
		{"", false, http.SameSiteLaxMode, true},
	}

	for _, tt := range tests {
		c := CookieConfiguration{SameSite: tt.setting, Secure: tt.secure}
		if got := c.SameSiteMode(); got != tt.want {
			t.Errorf("SameSiteMode(%q) = %v, want %v", tt.setting, got, tt.want)
		}
		if err := c.validate(); (err != nil) != tt.wantErr {
			t.Errorf("validate(%q, secure %v) = %v, want an error: %v", tt.setting, tt.secure, err, tt.wantErr)
		}
	}
}

func TestConfiguration_RateLimiterStore(t *testing.T) {
	file := writeFile(t, `
api_rate_limiter:
  limits:
    authentication: { requests: 3, seconds: 30 }
  redis:
    address: localhost:6379
    password: !env RATE_LIMITER_PASSWORD
`)
	t.Setenv("RATE_LIMITER_PASSWORD", "secret")

	c := NewConfiguration()
	if err := c.Initialize(file, environment.Test, version.Final); err != nil {
		t.Fatal(err)
	}

	if got := c.APIRateLimiter.Limits.Authentication; got.Requests != 3 || got.Seconds != 30 {
		t.Errorf("authentication = %+v", got)
	}
	if got := c.APIRateLimiter.Limits.Registration.Requests; got != 5 {
		t.Errorf("registration requests = %d, want the default", got)
	}

	t.Run("present", func(t *testing.T) {
		settings := struct{ Address, Password, Prefix string }{Prefix: "default"}
		if err := c.RateLimiterStore("redis", &settings); err != nil {
			t.Fatal(err)
		}
		if settings.Address != "localhost:6379" || settings.Password != "secret" || settings.Prefix != "default" {
			t.Errorf("settings = %+v", settings)
		}
	})

	t.Run("absent", func(t *testing.T) {
		settings := struct{ Address string }{Address: "default"}
		if err := c.RateLimiterStore("mongodb", &settings); err != nil || settings.Address != "default" {
			t.Errorf("settings = %+v, err = %v", settings, err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		settings := struct{ Address int }{}
		if err := c.RateLimiterStore("redis", &settings); err == nil {
			t.Error("a string must not decode into an int")
		}
	})
}

func TestConfiguration_Initialize_RemovedRateLimits(t *testing.T) {
	file := writeFile(t, "rate_limits:\n  authentication: { requests: 3, seconds: 30 }\n")

	if err := NewConfiguration().Initialize(file, environment.Test, version.Final); err == nil {
		t.Error("the section rate_limits must be refused")
	}
}
