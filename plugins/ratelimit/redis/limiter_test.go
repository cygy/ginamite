package redis

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func newTestClient(t *testing.T) (*goredis.Client, *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return client, server
}

func TestRateLimiter_Allow(t *testing.T) {
	client, server := newTestClient(t)
	limiter := NewRateLimiters(client, "test:")("authentication", 3, time.Minute)

	for i := 1; i <= 3; i++ {
		ok, retry, err := limiter.Allow(t.Context(), "1.2.3.4")
		if err != nil || !ok || retry != 0 {
			t.Fatalf("request %d: Allow() = %v, %v, %v", i, ok, retry, err)
		}
	}

	ok, retry, err := limiter.Allow(t.Context(), "1.2.3.4")
	if err != nil || ok || retry <= 0 || retry > time.Minute {
		t.Fatalf("over the limit: Allow() = %v, %v, %v", ok, retry, err)
	}

	if ttl := server.TTL("test:authentication:1.2.3.4"); ttl <= 0 || ttl > time.Minute {
		t.Errorf("ttl = %v, the counter must expire with its window", ttl)
	}

	// Another client and another group have their own counters.
	if ok, _, _ := limiter.Allow(t.Context(), "5.6.7.8"); !ok {
		t.Error("another client must not be limited")
	}
	other := NewRateLimiters(client, "test:")("registration", 3, time.Minute)
	if ok, _, _ := other.Allow(t.Context(), "1.2.3.4"); !ok {
		t.Error("another group must not be limited")
	}

	// The window ends: the counter expires.
	server.FastForward(time.Minute + time.Second)
	if ok, _, err := limiter.Allow(t.Context(), "1.2.3.4"); err != nil || !ok {
		t.Errorf("after the window: Allow() = %v, %v", ok, err)
	}
}

func TestRateLimiter_Allow_Error(t *testing.T) {
	client, server := newTestClient(t)
	limiter := NewRateLimiters(client, "")("authentication", 3, time.Minute)

	server.Close()

	ok, _, err := limiter.Allow(t.Context(), "1.2.3.4")
	if err == nil || ok {
		t.Errorf("Allow() = %v, %v, want the error of the storage", ok, err)
	}
}

func TestRateLimiter_Allow_NoExpiration(t *testing.T) {
	client, server := newTestClient(t)
	limiter := NewRateLimiters(client, "")("authentication", 1, time.Minute)

	// A counter which lost its expiration must not block the client forever.
	if err := server.Set("authentication:1.2.3.4", "5"); err != nil {
		t.Fatal(err)
	}

	ok, retry, err := limiter.Allow(t.Context(), "1.2.3.4")
	if err != nil || ok || retry != time.Minute {
		t.Errorf("Allow() = %v, %v, %v", ok, retry, err)
	}
}
