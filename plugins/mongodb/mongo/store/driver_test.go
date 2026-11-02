package store

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// These tests run against a server that can not be reached: they check that every
// method reaches the driver and reports its error. The success paths need a MongoDB server.

func offline(t *testing.T) Database {
	t.Helper()

	client, err := mongo.Connect(options.Client().
		SetHosts([]string{"127.0.0.1:1"}).
		SetServerSelectionTimeout(20 * time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	return NewDatabase(client.Database("app"))
}

func TestNewDatabase_Nil(t *testing.T) {
	if db := NewDatabase(nil); db != nil {
		t.Errorf("NewDatabase(nil) = %v, want a nil interface", db)
	}
}

func TestDatabase_Name(t *testing.T) {
	if got := offline(t).Name(); got != "app" {
		t.Errorf("Name() = %q", got)
	}
}

func TestDatabase_CreateCollection(t *testing.T) {
	if err := offline(t).CreateCollection(context.Background(), "users"); err == nil {
		t.Error("expected the connection error")
	}
}

func TestCollection_Unreachable(t *testing.T) {
	ctx := context.Background()
	c := offline(t).Collection("users")
	filter := bson.M{"a": 1}
	update := bson.M{"$set": filter}

	tests := []struct {
		name string
		call func() error
	}{
		{"InsertOne", func() error { _, err := c.InsertOne(ctx, filter); return err }},
		{"UpdateByID", func() error { _, err := c.UpdateByID(ctx, bson.NewObjectID(), update); return err }},
		{"UpdateOne", func() error { _, err := c.UpdateOne(ctx, filter, update); return err }},
		{"UpdateMany", func() error { _, err := c.UpdateMany(ctx, filter, update); return err }},
		{"DeleteOne", func() error { _, err := c.DeleteOne(ctx, filter); return err }},
		{"DeleteMany", func() error { _, err := c.DeleteMany(ctx, filter); return err }},
		{"FindOne", func() error { return c.FindOne(ctx, filter).Err() }},
		{"CountDocuments", func() error { _, err := c.CountDocuments(ctx, filter); return err }},
		{"Drop", func() error { return c.Drop(ctx) }},
		{"CreateOne", func() error {
			_, err := c.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "a", Value: 1}}})
			return err
		}},
		{"DropOne", func() error { return c.Indexes().DropOne(ctx, "a_1") }},
		{"DropAll", func() error { return c.Indexes().DropAll(ctx) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Error("expected the connection error")
			}
		})
	}
}

func TestCollection_Cursors(t *testing.T) {
	ctx := context.Background()
	c := offline(t).Collection("users")

	// The cursor must be a nil interface on failure, so that callers can compare it with nil.
	cursor, err := c.Find(ctx, bson.M{})
	if err == nil || cursor != nil {
		t.Errorf("Find = %v, %v", cursor, err)
	}

	cursor, err = c.Aggregate(ctx, bson.A{})
	if err == nil || cursor != nil {
		t.Errorf("Aggregate = %v, %v", cursor, err)
	}
}
