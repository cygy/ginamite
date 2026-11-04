package ratelimit

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

func newLimiter(string, int, time.Duration) request.RateLimiter { return nil }

func TestRegister(t *testing.T) {
	s := ginamite.NewServer()

	if err := Register(s, "redis", newLimiter); err != nil {
		t.Fatal(err)
	}
	if s.API.Hooks.NewRateLimiter == nil || s.API.Hooks.RateLimiterStore != "redis" {
		t.Errorf("hooks = %+v", s.API.Hooks)
	}
}

func TestRegister_Twice(t *testing.T) {
	s := ginamite.NewServer()
	if err := Register(s, "redis", newLimiter); err != nil {
		t.Fatal(err)
	}

	err := Register(s, "mongodb", newLimiter)
	if err == nil || !strings.Contains(err.Error(), "redis") || !strings.Contains(err.Error(), "mongodb") {
		t.Errorf("Register() = %v, want an error naming the two storages", err)
	}
	if s.API.Hooks.RateLimiterStore != "redis" {
		t.Errorf("store = %q, the first plugin must stay", s.API.Hooks.RateLimiterStore)
	}
}
