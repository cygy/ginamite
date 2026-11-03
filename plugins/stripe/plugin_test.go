package stripe

import (
	"testing"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/config/version"

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
			"missing func":   {Store: newFakeStore(), SettingsFunc: func(*config.Configuration) (Settings, error) { return testSettings(), nil }},
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
