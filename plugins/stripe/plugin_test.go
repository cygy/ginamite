package stripe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"

	"github.com/cygy/ginamite/api/route"

	"github.com/gin-gonic/gin"
)

func succeed(*gin.Context, Succeeded) error { return nil }

func TestPlugin_Configure(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		p := NewPlugin(newFakeStore(), succeed)
		if err := p.Configure(ginamite.NewServer()); err != nil || p.Service() != nil {
			t.Errorf("Configure() = %v, service = %v, want no service", err, p.Service())
		}
	})

	t.Run("from the section of the configuration", func(t *testing.T) {
		s := ginamite.NewServer()
		s.Config = configuration(t, "stripe: {enabled: true, secret_api_key: sk, secret_webhook_key: whsec, currency: usd}\n")

		p := NewPlugin(newFakeStore(), succeed)
		if err := p.Configure(s); err != nil {
			t.Fatal(err)
		}
		if p.Service() == nil || p.Service().Settings().Currency != "usd" || p.Service().live {
			t.Errorf("service = %+v, want a test mode service", p.Service())
		}
	})

	t.Run("from SettingsFunc, live in the final production", func(t *testing.T) {
		s := ginamite.NewServer()
		s.Config.Environment, s.Config.Version = environment.Production, version.Final

		p := NewPlugin(newFakeStore(), succeed)
		p.SettingsFunc = func(*config.Configuration) (Settings, error) { return testSettings(), nil }
		if err := p.Configure(s); err != nil {
			t.Fatal(err)
		}
		if p.Service() == nil || !p.Service().live {
			t.Errorf("service = %+v, want a live service", p.Service())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		tests := map[string]*Plugin{
			"settings error": {Store: newFakeStore(), Succeeded: succeed, SettingsFunc: func(*config.Configuration) (Settings, error) { return Settings{}, errTest }},
			"missing secret": {Store: newFakeStore(), Succeeded: succeed, SettingsFunc: func(*config.Configuration) (Settings, error) { return Settings{Enabled: true}, nil }},
			"missing store":  {Succeeded: succeed, SettingsFunc: func(*config.Configuration) (Settings, error) { return testSettings(), nil }},
			"missing order": {Store: newFakeStore(), Succeeded: succeed, SettingsFunc: func(*config.Configuration) (Settings, error) {
				settings := testSettings()
				settings.IntentEnabled = true
				return settings, nil
			}},
			"missing func": {Store: newFakeStore(), SettingsFunc: func(*config.Configuration) (Settings, error) { return testSettings(), nil }},
		}

		for name, p := range tests {
			t.Run(name, func(t *testing.T) {
				if err := p.Configure(ginamite.NewServer()); err == nil || p.Service() != nil {
					t.Errorf("Configure() = %v, want an error and no service", err)
				}
			})
		}
	})
}

func TestPlugin_Configure_WebServer(t *testing.T) {
	s := ginamite.NewServer()
	s.Config = configuration(t, "stripe: {enabled: true, publishable_api_key: pk_test, intent_enabled: true}\n")

	// A web server has neither the secrets, nor the store, nor the order function.
	p := &Plugin{}
	if err := p.configure(s, ginamite.RoleWeb); err != nil {
		t.Fatalf("configure() = %v", err)
	}
	if p.Service() != nil {
		t.Error("the web server must not have a service")
	}
	if vars := p.WebVars(); vars[VarPublishableKey] != "pk_test" || vars[VarIntentEnabled] != true || vars[VarCheckoutEnabled] != false {
		t.Errorf("WebVars() = %v", vars)
	}

	// The API needs them.
	if err := (&Plugin{}).configure(s, ginamite.RoleAPI); err == nil {
		t.Error("configure() = nil, want an error for the API without secrets")
	}
}

func TestPlugin_WebVars(t *testing.T) {
	t.Run("not configured", func(t *testing.T) {
		vars := NewPlugin(newFakeStore(), succeed).WebVars()
		if vars[VarPublishableKey] != "" || vars[VarIntentEnabled] != false || vars[VarCheckoutEnabled] != false {
			t.Errorf("WebVars() = %v", vars)
		}
	})

	t.Run("not enabled", func(t *testing.T) {
		s := ginamite.NewServer()
		s.Config = configuration(t, "stripe: {publishable_api_key: pk_test, intent_enabled: true}\n")

		p := NewPlugin(newFakeStore(), succeed)
		if err := p.Configure(s); err != nil {
			t.Fatal(err)
		}
		if vars := p.WebVars(); vars[VarIntentEnabled] != false || vars[VarCheckoutEnabled] != false {
			t.Errorf("WebVars() = %v, want nothing enabled", vars)
		}
	})

	t.Run("enabled", func(t *testing.T) {
		s := ginamite.NewServer()
		s.Config = configuration(t, "stripe: {enabled: true, secret_api_key: sk, secret_webhook_key: whsec, publishable_api_key: pk_test, checkout_enabled: true}\n")

		p := NewPlugin(newFakeStore(), succeed)
		p.OrderFunc = testOrder
		if err := p.Configure(s); err != nil {
			t.Fatal(err)
		}
		if vars := p.WebVars(); vars[VarPublishableKey] != "pk_test" || vars[VarIntentEnabled] != false || vars[VarCheckoutEnabled] != true {
			t.Errorf("WebVars() = %v", vars)
		}
	})
}

func TestPlugin_addRoutes(t *testing.T) {
	var calls []string
	guard := func(name string) gin.HandlerFunc {
		return func(c *gin.Context) { calls = append(calls, name); c.Next() }
	}
	// The middleware of the terms is missing: it is skipped.
	handlers := route.Middlewares{
		MustBeAuthenticated:       guard("authenticated"),
		MustAuthenticationBeValid: guard("valid"),
	}

	paths := func(settings Settings) map[string]bool {
		p := NewPlugin(newFakeStore(), succeed)
		p.OrderFunc = func(c *gin.Context) (Order, bool) {
			calls = append(calls, "order")
			c.AbortWithStatus(http.StatusPreconditionFailed)
			return Order{}, false
		}
		p.service = NewService(settings, p.Store, p.Succeeded, false)

		gin.SetMode(gin.TestMode)
		router := gin.New()
		p.addRoutes(router.Group("/v1"), handlers)

		found := map[string]bool{}
		for _, r := range router.Routes() {
			found[r.Method+" "+r.Path] = true
		}

		calls = nil
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/stripe/intent", nil)
		router.ServeHTTP(httptest.NewRecorder(), request)

		return found
	}

	t.Run("webhook only", func(t *testing.T) {
		found := paths(testSettings())
		if len(found) != 1 || !found["POST /v1/stripe/hook"] {
			t.Errorf("routes = %v", found)
		}
	})

	t.Run("intent and checkout behind the guards", func(t *testing.T) {
		settings := testSettings()
		settings.IntentEnabled, settings.CheckoutEnabled = true, true
		settings.CheckoutPath = "/store/packs/:id/checkout"

		found := paths(settings)
		if len(found) != 3 || !found["POST /v1/stripe/intent"] || !found["POST /v1/store/packs/:id/checkout"] {
			t.Errorf("routes = %v", found)
		}
		if len(calls) != 3 || calls[0] != "authenticated" || calls[1] != "valid" || calls[2] != "order" {
			t.Errorf("calls = %v, want the guards then the order", calls)
		}
	})
}
