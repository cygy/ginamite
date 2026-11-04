// Package stripe takes payments with Stripe: it creates the checkout sessions and the payment intents, saves the
// payments in a Store, and handles the events of the webhook, so that the application delivers what was paid for once
// per payment, see SucceededFunc.
//
// The application chooses what is sold and what a payment delivers; the plugin handles Stripe. The plugin adds to the API
// the route of the webhook, and the routes creating the payment intents and the checkout sessions when the settings
// enable them: they ask the application what is bought with Plugin.OrderFunc, and the plugin builds the URLs where
// Stripe sends the user back from the web routes of the server. The pages of the website read the settings they need
// with Plugin.WebVars. The package mongostore provides a Store saving the payments in MongoDB.
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

	// OrderFunc returns what a request of the routes creating the payments buys, see OrderFunc. It is required if the
	// settings enable the intents or the checkouts.
	OrderFunc OrderFunc

	// SettingsFunc returns the settings, read from the configuration of the application for instance. The settings
	// are read from the section "stripe" of the configuration of the server if it is nil.
	SettingsFunc func(c *config.Configuration) (Settings, error)

	// ClientOptions are the options of the client of the API of Stripe, meant for the tests.
	ClientOptions []stripeapi.ClientOption

	service *Service
	// loaded are the settings read by Configure, with their default values.
	loaded Settings
}

var _ ginamite.Plugin = (*Plugin)(nil)

// NewPlugin returns a Plugin saving the payments in the store, and delivering them with succeeded.
func NewPlugin(store Store, succeeded SucceededFunc) *Plugin {
	return &Plugin{Store: store, Succeeded: succeeded}
}

// Configure builds the Service and adds the routes to the API, if the settings enable the plugin (the web server only
// reads the settings, see WebVars): the webhook, then the
// routes creating the payment intents and the checkout sessions, behind the middlewares checking that the user is
// authenticated and has accepted the latest terms. The payments are real ones only in the final version of the
// production. It returns an error if the settings are invalid, or if the store or the SucceededFunc is missing, or the
// OrderFunc when the settings enable the intents or the checkouts.
func (p *Plugin) Configure(s *ginamite.Server) error {
	return p.configure(s, s.Role())
}

// configure configures the plugin for the server, running as role.
func (p *Plugin) configure(s *ginamite.Server, role ginamite.Role) error {
	settings, err := p.settings(s.Config)
	if err != nil {
		return fmt.Errorf("invalid stripe configuration: %w", err)
	}
	p.loaded = settings.withDefaults()
	if !settings.Enabled {
		return nil
	}

	// The web server only reads the settings, see WebVars: it creates no payment and needs none of the secrets.
	if role == ginamite.RoleWeb {
		return nil
	}

	if err := settings.validate(); err != nil {
		return fmt.Errorf("invalid stripe configuration: %w", err)
	}
	if p.Store == nil || p.Succeeded == nil {
		return fmt.Errorf("the stripe plugin needs a store and a function delivering the payments")
	}

	if (settings.IntentEnabled || settings.CheckoutEnabled) && p.OrderFunc == nil {
		return fmt.Errorf("the stripe plugin needs an OrderFunc to create the payment intents and the checkout sessions")
	}

	live := s.Config.IsProductionEnvironment() && s.Config.IsFinalVersion()
	p.service = NewService(settings, p.Store, p.Succeeded, live, p.ClientOptions...)

	s.API.AddCustomRoutes(func(group *gin.RouterGroup, handlers route.Middlewares, _ bool, _ func(bool, []string) gin.HandlerFunc) {
		p.addRoutes(group, handlers)
	})

	return nil
}

// Service returns the service creating the payments, or nil if the plugin is not enabled, not yet
// configured, or running in the web server.
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

// Template variables given by WebVars.
const (
	// VarPublishableKey is the name of the variable holding the publishable key, for Stripe.js.
	VarPublishableKey = "StripePublishableKey"
	// VarIntentEnabled is the name of the variable telling whether the payment intents are enabled.
	VarIntentEnabled = "IntentEnabled"
	// VarCheckoutEnabled is the name of the variable telling whether the checkout sessions are enabled.
	VarCheckoutEnabled = "CheckoutEnabled"
)

// WebVars returns the variables the pages paying with Stripe need, from the settings of the plugin: the publishable
// key (VarPublishableKey), and whether the payment intents (VarIntentEnabled) and the checkout sessions
// (VarCheckoutEnabled) are enabled. The plugin does not render the pages, which belong to the application: it sets
// them in the web handler, with page.SetVar for instance. The variables are those of a plugin which is not enabled
// (empty key, both false) before Configure, and when the plugin is not enabled. The web server reads them from the section
// "stripe" of its own configuration, which needs only "enabled", "publishable_api_key", "intent_enabled" and
// "checkout_enabled": it does not need the secrets, which are not validated there.
func (p *Plugin) WebVars() map[string]any {
	enabled := p.loaded.Enabled

	return map[string]any{
		VarPublishableKey:  p.loaded.PublishableAPIKey,
		VarIntentEnabled:   enabled && p.loaded.IntentEnabled,
		VarCheckoutEnabled: enabled && p.loaded.CheckoutEnabled,
	}
}

// nonNil returns the handlers which are not nil.
func nonNil(handlers ...gin.HandlerFunc) []gin.HandlerFunc {
	kept := make([]gin.HandlerFunc, 0, len(handlers))
	for _, h := range handlers {
		if h != nil {
			kept = append(kept, h)
		}
	}

	return kept
}

// withGuards returns the guards followed by the handler, in a new slice.
func withGuards(guards []gin.HandlerFunc, handler gin.HandlerFunc) []gin.HandlerFunc {
	return append(append(make([]gin.HandlerFunc, 0, len(guards)+1), guards...), handler)
}

// addRoutes adds the routes of the service to the group of the version of the API: the webhook, then the routes
// creating the payments the settings enable, behind the middlewares checking that the user is authenticated and has
// accepted the latest terms, before the application looks at the order.
func (p *Plugin) addRoutes(group *gin.RouterGroup, handlers route.Middlewares) {
	settings := p.service.Settings()
	group.POST(settings.WebhookPath, p.service.Webhook)

	guards := nonNil(handlers.MustBeAuthenticated, handlers.MustAuthenticationBeValid, handlers.MustHaveAcceptedTheLatestVersionOfTerms)
	if settings.IntentEnabled {
		group.POST(settings.IntentPath, withGuards(guards, p.service.IntentHandler(p.OrderFunc))...)
	}
	if settings.CheckoutEnabled {
		group.POST(settings.CheckoutPath, withGuards(guards, p.service.CheckoutHandler(p.OrderFunc))...)
	}
}
