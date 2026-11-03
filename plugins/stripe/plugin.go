// Package stripe takes payments with Stripe: it creates the checkout sessions and the payment intents, saves the
// payments in a Store, and handles the events of the webhook, so that the application delivers what was paid for once
// per payment, see SucceededFunc.
//
// The application chooses what is sold and what a payment delivers; the plugin handles Stripe. Its routes creating
// the payments call Plugin.Service, and the plugin adds the route of the webhook to the API. The package mongostore
// provides a Store saving the payments in MongoDB.
package stripe

import (
	"fmt"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
	stripeapi "github.com/stripe/stripe-go/v87"
)

// Plugin adds the payments with Stripe to the servers: it builds the Service from the settings, and adds the route of
// its webhook to the API. It does nothing if the settings do not enable it.
type Plugin struct {
	// Store saves the payments, and SucceededFunc delivers them: both are required.
	Store     Store
	Succeeded SucceededFunc

	// SettingsFunc returns the settings, read from the configuration of the application for instance. The settings
	// are read from the section "stripe" of the configuration of the server if it is nil.
	SettingsFunc func(c *config.Configuration) (Settings, error)

	// ClientOptions are the options of the client of the API of Stripe, meant for the tests.
	ClientOptions []stripeapi.ClientOption

	service *Service
}

var _ ginamite.Plugin = (*Plugin)(nil)

// NewPlugin returns a Plugin saving the payments in the store, and delivering them with succeeded.
func NewPlugin(store Store, succeeded SucceededFunc) *Plugin {
	return &Plugin{Store: store, Succeeded: succeeded}
}

// Configure builds the Service and adds the webhook to the API, if the settings enable the plugin. The payments are
// real ones only in the final version of the production. It returns an error if the settings are invalid, or if the
// store or the SucceededFunc is missing.
func (p *Plugin) Configure(s *ginamite.Server) error {
	settings, err := p.settings(s.Config)
	if err != nil {
		return fmt.Errorf("invalid stripe configuration: %w", err)
	}
	if !settings.Enabled {
		return nil
	}

	if err := settings.validate(); err != nil {
		return fmt.Errorf("invalid stripe configuration: %w", err)
	}
	if p.Store == nil || p.Succeeded == nil {
		return fmt.Errorf("the stripe plugin needs a store and a function delivering the payments")
	}

	live := s.Config.IsProductionEnvironment() && s.Config.IsFinalVersion()
	p.service = NewService(settings, p.Store, p.Succeeded, live, p.ClientOptions...)

	service := p.service
	s.API.AddCustomRoutes(func(group *gin.RouterGroup, _ route.Middlewares, _ bool, _ func(bool, []string) gin.HandlerFunc) {
		group.POST(service.Settings().WebhookPath, service.Webhook)
	})

	return nil
}

// Service returns the service creating the payments, or nil if the plugin is not enabled, or not yet configured.
func (p *Plugin) Service() *Service {
	return p.service
}

// settings returns the settings of the plugin.
func (p *Plugin) settings(c *config.Configuration) (Settings, error) {
	if p.SettingsFunc != nil {
		return p.SettingsFunc(c)
	}

	return settingsFrom(c)
}
