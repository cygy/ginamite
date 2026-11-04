package redis

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

func loadConfiguration(t *testing.T, content string) *config.Configuration {
	t.Helper()

	path := filepath.Join(t.TempDir(), "app.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	c := config.NewConfiguration()
	if err := c.Initialize(path, "test", "final"); err != nil {
		t.Fatal(err)
	}

	return c
}

func serverConfiguration(t *testing.T, server *miniredis.Miniredis, extra string) *config.Configuration {
	t.Helper()

	return loadConfiguration(t, "api_rate_limiter:\n  redis:\n    host: "+server.Host()+"\n    port: "+strconv.Itoa(server.Server().Addr().Port)+"\n"+extra)
}

func TestPlugin_Connect(t *testing.T) {
	server := miniredis.RunT(t)

	p := NewPlugin()
	if err := p.Connect(serverConfiguration(t, server, "    prefix: \"app:\"\n")); err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	if p.settings.Prefix != "app:" || p.settings.Timeout != 2 {
		t.Errorf("settings = %+v", p.settings)
	}
}

func TestPlugin_Connect_Password(t *testing.T) {
	server := miniredis.RunT(t)
	server.RequireAuth("secret")
	t.Setenv("RATE_LIMITER_REDIS_PASSWORD", "secret")

	p := NewPlugin()
	if err := p.Connect(serverConfiguration(t, server, "    password: !env RATE_LIMITER_REDIS_PASSWORD\n")); err != nil {
		t.Fatal(err)
	}
	p.Close()

	if err := NewPlugin().Connect(serverConfiguration(t, server, "")); err == nil {
		t.Error("Connect() must fail without the password")
	}
}

func TestPlugin_Connect_Errors(t *testing.T) {
	t.Run("missing host", func(t *testing.T) {
		err := NewPlugin().Connect(loadConfiguration(t, "api_rate_limiter:\n  limits: {}\n"))
		if err == nil || !strings.Contains(err.Error(), "api_rate_limiter.redis.host") {
			t.Errorf("Connect() = %v", err)
		}
	})

	t.Run("invalid settings", func(t *testing.T) {
		if err := NewPlugin().Connect(loadConfiguration(t, "api_rate_limiter:\n  redis:\n    port: abc\n")); err == nil {
			t.Error("Connect() must refuse a port which is not a number")
		}
	})

	t.Run("unreachable server", func(t *testing.T) {
		server := miniredis.RunT(t)
		c := serverConfiguration(t, server, "    timeout: 1\n")
		server.Close()

		if err := NewPlugin().Connect(c); err == nil {
			t.Error("Connect() must fail when the server is not reachable")
		}
	})
}

func TestPlugin_Close_NotConnected(t *testing.T) {
	NewPlugin().Close()
}

func TestPlugin_Configure_NotConnected(t *testing.T) {
	s := ginamite.NewServer()
	if err := NewPlugin().Configure(s); err != nil {
		t.Fatal(err)
	}
	if s.API.Hooks.NewRateLimiter != nil {
		t.Error("a plugin which is not connected must not set the limiters")
	}
}

func TestPlugin_Configure(t *testing.T) {
	server := miniredis.RunT(t)
	p := NewPlugin()
	if err := p.Connect(serverConfiguration(t, server, "")); err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	s := ginamite.NewServer()
	if err := p.Configure(s); err != nil {
		t.Fatal(err)
	}
	if s.API.Hooks.NewRateLimiter == nil || s.API.Hooks.RateLimiterStore != Store {
		t.Errorf("hooks = %+v", s.API.Hooks)
	}
	if err := p.Configure(s); err == nil {
		t.Error("a second rate limiting plugin must be refused")
	}
}
