package functions

import (
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

func TestTerms_UpdateVersionOfTermsAcceptedByUser(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	terms := NewTerms(func() store.Database { return db })

	if err := terms.UpdateVersionOfTermsAcceptedByUser(ctx, id.Hex(), "v3"); err != nil {
		t.Fatal(err)
	}
	checkUpdate(t, db, model.UserCollection, id, bson.M{"$set": bson.M{"versionOfTermsAccepted": "v3"}})

	if err := terms.UpdateVersionOfTermsAcceptedByUser(ctx, "nope", "v3"); err == nil {
		t.Error("an invalid user ID must be rejected")
	}
}

func TestTerms_GetLatestVersionOfTermsAcceptedByUser(t *testing.T) {
	id := bson.NewObjectID()
	db := mock.NewDatabase("app")
	db.Coll(model.UserCollection).One = mock.Result{Value: bson.M{"_id": id, "versionOfTermsAccepted": "v2"}}
	terms := NewTerms(func() store.Database { return db })

	if got, err := terms.GetLatestVersionOfTermsAcceptedByUser(ctx, id.Hex()); err != nil || got != "v2" {
		t.Errorf("got %q, %v, want v2", got, err)
	}

	t.Run("unknown user", func(t *testing.T) {
		db := mock.NewDatabase("app")
		terms := NewTerms(func() store.Database { return db })

		if got, err := terms.GetLatestVersionOfTermsAcceptedByUser(ctx, id.Hex()); err != nil || got != "" {
			t.Errorf("got %q, %v, want an empty version", got, err)
		}
	})

	t.Run("database error", func(t *testing.T) {
		db := mock.NewDatabase("app")
		db.Coll(model.UserCollection).Err = errBoom
		terms := NewTerms(func() store.Database { return db })

		if _, err := terms.GetLatestVersionOfTermsAcceptedByUser(ctx, id.Hex()); !errors.Is(err, errBoom) {
			t.Errorf("err = %v, want %v", err, errBoom)
		}
	})
}
