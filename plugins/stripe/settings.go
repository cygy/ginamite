package stripe

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/common/config"
)

// Section is the name of the section of the configuration holding the settings, see Settings.
const Section = "stripe"

// DefaultWebhookPath is the path of the route receiving the events of Stripe, relative to the group of the version of
// the API.
const DefaultWebhookPath = "/stripe/hook"

// DefaultCurrency is the currency of the payments when the settings give none.
const DefaultCurrency = "eur"

// Settings holds the settings of Stripe, in the section "stripe" of the configuration.
type Settings struct {
	// Enabled enables the plugin: without it, the plugin adds no route and Plugin.Service returns nil.
	Enabled bool

	// PublishableAPIKey is the key given to the browsers, which the plugin does not use.
	PublishableAPIKey string `yaml:"publishable_api_key"`
	// SecretAPIKey is the key of the calls to the API of Stripe.
	SecretAPIKey string `yaml:"secret_api_key"`
	// SecretWebhookKey is the secret checking the signature of the events sent to the webhook.
	SecretWebhookKey string `yaml:"secret_webhook_key"`

	// Currency is the ISO code of the currency of the payments, in lower case ("eur", "usd"). DefaultCurrency if empty.
	Currency string
	// PaymentMethodTypes are the methods accepted by the checkout sessions ("card", "paypal"). Stripe chooses them
	// from the settings of the account if empty.
	PaymentMethodTypes []string `yaml:"payment_method_types"`

	// WebhookPath is the path of the webhook, relative to the group of the version of the API. DefaultWebhookPath if
	// empty.
	WebhookPath string `yaml:"webhook_path"`
}

// settingsFrom returns the settings of the section "stripe" of the configuration.
func settingsFrom(c *config.Configuration) (Settings, error) {
	var settings Settings
	err := c.Section(Section, &settings)

	return settings, err
}

// withDefaults returns the settings with the default values of the empty ones.
func (s Settings) withDefaults() Settings {
	if s.Currency == "" {
		s.Currency = DefaultCurrency
	}
	s.Currency = strings.ToLower(s.Currency)

	if s.WebhookPath == "" {
		s.WebhookPath = DefaultWebhookPath
	}

	return s
}

// validate returns an error if the settings miss a secret.
func (s Settings) validate() error {
	if s.SecretAPIKey == "" {
		return errors.New("secret_api_key is required")
	}
	if s.SecretWebhookKey == "" {
		return errors.New("secret_webhook_key is required")
	}

	return nil
}
