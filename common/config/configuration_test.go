package config

import (
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
	if c.Server.Host != "localhost" || c.Server.Port != 80 || c.DefaultLocale != "en-us" || c.JWT.SigningMethod != "HS256" {
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

	if c.RateLimits.Authentication.Requests != 10 || c.RateLimits.Authentication.Seconds != 60 {
		t.Errorf("authentication = %+v", c.RateLimits.Authentication)
	}
	if c.RateLimits.Registration.Requests == 0 || c.RateLimits.ForgotPassword.Requests == 0 {
		t.Errorf("limits = %+v", c.RateLimits)
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
