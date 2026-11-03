//go:build integration

package functions

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// TestCreateIndexes_Integration_Conflict creates, on a real MongoDB server reachable at $MONGO_HOST:$MONGO_PORT
// (localhost:27017 if unset), indexes whose name is used by an index of an older definition, and checks that they are
// replaced. Run it with "go test -tags=integration ./plugins/mongodb/...", or "make integration".
func TestCreateIndexes_Integration_Conflict(t *testing.T) {
	host, port := os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "27017"
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port)).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatalf("Connect() = %v", err)
	}
	t.Cleanup(func() { client.Disconnect(context.Background()) }) //nolint:errcheck // best-effort cleanup

	db := client.Database("ginamite_integration_test_" + uuid.NewString()[:8])
	t.Cleanup(func() { db.Drop(context.Background()) }) //nolint:errcheck // best-effort cleanup

	// The old indexes: a partial filter on another field (other specification), and no unique option (other options).
	old := []mongo.IndexModel{
		{Keys: bson.D{{Key: "google", Value: 1}}, Options: options.Index().SetName("google_1_1").SetUnique(true).
			SetPartialFilterExpression(bson.M{"facebook": bson.M{"$exists": true}})},
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetName("email_1_1")},
	}
	if _, err := db.Collection("users").Indexes().CreateMany(ctx, old); err != nil {
		t.Fatalf("CreateMany() = %v", err)
	}

	fns := New(func() store.Database { return store.NewDatabase(db) })
	fns.CreateIndexes(ctx, []IndexToCreate{
		{Collection: "users", Index: mongo.IndexModel{Keys: bson.D{{Key: "google", Value: 1}}, Options: options.Index().
			SetName("google_1_1").SetUnique(true).SetPartialFilterExpression(bson.M{"google": bson.M{"$exists": true}})}},
		{Collection: "users", Index: mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().
			SetName("email_1_1").SetUnique(true)}},
	})

	cursor, err := db.Collection("users").Indexes().List(ctx)
	if err != nil {
		t.Fatalf("List() = %v", err)
	}
	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		t.Fatalf("All() = %v", err)
	}
	byName := map[string]bson.M{}
	for _, index := range indexes {
		byName[index["name"].(string)] = index
	}

	// The nested documents are decoded as bson.D.
	filter, _ := byName["google_1_1"]["partialFilterExpression"].(bson.D)
	if len(filter) != 1 || filter[0].Key != "google" {
		t.Errorf("partial filter of google_1_1 = %v, want it on google", filter)
	}
	if unique, _ := byName["email_1_1"]["unique"].(bool); !unique {
		t.Errorf("email_1_1 = %v, want it unique", byName["email_1_1"])
	}
}
