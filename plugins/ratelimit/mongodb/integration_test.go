//go:build integration

package mongodb

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cygy/ginamite"

	"github.com/google/uuid"
)

// TestPlugin_Integration_Allow connects to a real MongoDB server, reachable at $MONGO_HOST:$MONGO_PORT (localhost:27017
// if unset), creates the collection of the counters and counts requests. Run it with
// "go test -tags=integration ./plugins/ratelimit/...", or "make integration".
func TestPlugin_Integration_Allow(t *testing.T) {
	port := 27017
	if value, err := strconv.Atoi(os.Getenv("MONGO_PORT")); err == nil {
		port = value
	}
	host := os.Getenv("MONGO_HOST")
	if host == "" {
		host = "localhost"
	}

	collection := "rateLimits_" + uuid.NewString()
	configuration := loadConfiguration(t, "api_rate_limiter:\n  mongodb:\n    host: "+host+"\n    port: "+strconv.Itoa(port)+
		"\n    database: ginamite_integration_test\n    collection: "+collection+"\n")

	p := NewPlugin()
	if err := p.Connect(configuration); err != nil {
		t.Fatalf("Connect() = %v", err)
	}
	t.Cleanup(p.Close)
	t.Cleanup(func() {
		_ = p.session.Client.Database("ginamite_integration_test").Collection(collection).Drop(t.Context())
	})

	s := ginamite.NewServer()
	if err := p.Configure(s); err != nil {
		t.Fatalf("Configure() = %v", err)
	}
	// Idempotent: the collection and the index exist.
	if err := prepare(t.Context(), p.database, collection); err != nil {
		t.Fatalf("prepare() = %v", err)
	}

	limiter := s.API.Hooks.NewRateLimiter("authentication", 2, time.Minute)
	client := uuid.NewString()
	for i := 1; i <= 3; i++ {
		ok, _, err := limiter.Allow(t.Context(), client)
		if err != nil || ok != (i <= 2) {
			t.Fatalf("request %d: Allow() = %v, %v", i, ok, err)
		}
	}
}
