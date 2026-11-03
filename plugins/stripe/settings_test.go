package stripe

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cygy/ginamite/common/config"
)

func TestSettings_WithDefaults(t *testing.T) {
	if s := (Settings{}).withDefaults(); s.Currency != DefaultCurrency || s.WebhookPath != DefaultWebhookPath {
		t.Errorf("defaults = %+v", s)
	}

	s := Settings{Currency: "USD", WebhookPath: "/payments/hook"}.withDefaults()
	if s.Currency != "usd" || s.WebhookPath != "/payments/hook" {
		t.Errorf("settings = %+v, want the given values, the currency in lower case", s)
	}
}

func TestSettings_Validate(t *testing.T) {
	if err := testSettings().validate(); err != nil {
		t.Errorf("validate() = %v", err)
	}
	if err := (Settings{SecretWebhookKey: "whsec"}).validate(); err == nil {
		t.Error("validate() = nil, want an error without the API key")
	}
	if err := (Settings{SecretAPIKey: "sk"}).validate(); err == nil {
		t.Error("validate() = nil, want an error without the webhook secret")
	}
}

// configuration returns a configuration loaded from the content of a file.
func configuration(t *testing.T, content string) *config.Configuration {
	t.Helper()

	file := filepath.Join(t.TempDir(), "server.yaml")
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	c := config.NewConfiguration()
	if err := c.Initialize(file, "test", "final"); err != nil {
		t.Fatal(err)
	}

	return c
}

func TestSettingsFrom(t *testing.T) {
	c := configuration(t, "stripe:\n  enabled: true\n  secret_api_key: sk\n  secret_webhook_key: whsec\n  currency: usd\n  payment_method_types: [card]\n")

	s, err := settingsFrom(c)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Enabled || s.SecretAPIKey != "sk" || s.SecretWebhookKey != "whsec" || s.Currency != "usd" || len(s.PaymentMethodTypes) != 1 {
		t.Errorf("settings = %+v", s)
	}

	if _, err := settingsFrom(configuration(t, "stripe: [not, a, map]\n")); err == nil {
		t.Error("settingsFrom() = nil, want an error for an invalid section")
	}
}
