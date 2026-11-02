//go:build integration

package mongodb

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestPlugin_Integration_Connect connects to a real MongoDB server, reachable at $MONGO_HOST:$MONGO_PORT
// (localhost:27017 if unset), and checks that a document survives a round trip. Run it with
// "go test -tags=integration ./plugins/mongodb/...", or "make integration".
func TestPlugin_Integration_Connect(t *testing.T) {
	configuration := newConfiguration(Service)
	configuration.Database.Host = envOr(t, "MONGO_HOST", "localhost")
	configuration.Database.Port = envIntOr(t, "MONGO_PORT", 27017)
	configuration.Database.Database = "ginamite_integration_test"
	configuration.Database.TimeOut = 10

	p := NewPlugin()
	if err := p.Connect(configuration); err != nil {
		t.Fatalf("Connect() = %v, want a session on the MongoDB server", err)
	}
	t.Cleanup(p.Close)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	collection := p.session.Client.Database(configuration.Database.Database).Collection("integration_test")
	id := uuid.NewString()
	t.Cleanup(func() { collection.DeleteOne(context.Background(), bson.M{"_id": id}) }) //nolint:errcheck // best-effort cleanup

	if _, err := collection.InsertOne(ctx, bson.M{"_id": id, "value": "ok"}); err != nil {
		t.Fatalf("InsertOne() = %v", err)
	}

	var document bson.M
	if err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&document); err != nil {
		t.Fatalf("FindOne() = %v", err)
	}
	if document["value"] != "ok" {
		t.Errorf("value = %v, want %q", document["value"], "ok")
	}
}

// envOr returns the value of the environment variable name, or fallback if it is unset.
func envOr(t *testing.T, name, fallback string) string {
	t.Helper()

	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

// envIntOr returns the value of the environment variable name as an integer, or fallback if it is unset or invalid.
func envIntOr(t *testing.T, name string, fallback int) int {
	t.Helper()

	if value := os.Getenv(name); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}

	return fallback
}
