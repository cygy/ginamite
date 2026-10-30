package functions

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestCreateCollections(t *testing.T) {
	db := withMainSession(t)
	withOptions := options.CreateCollection().SetCapped(true)

	CreateCollections(context.Background(), []CollectionToCreate{{Name: "plain"}, {Name: "configured", Options: withOptions}})

	if got := db.Created(); len(got) != 2 || got[0] != "plain" || got[1] != "configured" {
		t.Fatalf("created = %v", got)
	}
	// A nil builder must not be given to the driver, it would panic.
	if got := db.CreateOptions("plain"); len(got) != 0 {
		t.Errorf("options of plain = %v", got)
	}
	if got := db.CreateOptions("configured"); len(got) != 1 {
		t.Errorf("options of configured = %v", got)
	}

	// A failure is only logged: the other collections are created.
	db.CreateCollectionErr = errBoom
	CreateCollections(context.Background(), []CollectionToCreate{{Name: "a"}, {Name: "b"}})
	if got := db.Created(); len(got) != 4 {
		t.Errorf("created = %v", got)
	}
}

func TestRemoveCollections(t *testing.T) {
	ctx := context.Background()

	t.Run("drops the indexes and the collection", func(t *testing.T) {
		db := withMainSession(t)
		RemoveCollections(ctx, []string{"a", "b"})

		for _, name := range []string{"a", "b"} {
			if got := methods(db, name); len(got) != 1 || got[0] != "Drop" {
				t.Errorf("%s: calls = %v", name, got)
			}
			if db.Coll(name).IndexView == nil || db.Coll(name).IndexView.AllDrop != 1 {
				t.Errorf("%s: the indexes must be dropped", name)
			}
		}
	})

	t.Run("a collection that does not exist is not an error", func(t *testing.T) {
		db := withMainSession(t)
		db.Coll("a").Err = mongo.CommandError{Code: 26, Message: "ns not found"}
		db.Coll("b").Err = errBoom

		RemoveCollections(ctx, []string{"a", "b"})
		if len(methods(db, "a")) != 1 || len(methods(db, "b")) != 1 {
			t.Error("every collection must be processed")
		}
	})

	t.Run("failures to drop the indexes are only logged", func(t *testing.T) {
		db := withMainSession(t)
		db.Coll("a").Indexes()
		db.Coll("a").IndexView.Err = errBoom

		RemoveCollections(ctx, []string{"a"})
		if got := methods(db, "a"); len(got) != 1 {
			t.Errorf("calls = %v", got)
		}
	})
}

func TestRemoveUnusedCollections(t *testing.T) {
	db := withMainSession(t)
	RemoveUnusedCollections(context.Background())

	if len(db.Created()) != 0 {
		t.Errorf("created = %v", db.Created())
	}
}

func TestCreateAllCollections(t *testing.T) {
	db := withMainSession(t)
	CreateAllCollections(context.Background())

	want := []string{
		model.AuthTokenCollection, model.DeletedUserCollection, model.DisabledUserCollection, model.ForgotPasswordProcessCollection,
		model.IPLocationCollection, model.UpdatePropertyProcessCollection, model.UserCollection,
	}
	got := db.Created()
	if len(got) != len(want) {
		t.Fatalf("created = %v, want %v", got, want)
	}
	for i, name := range want {
		if got[i] != name {
			t.Errorf("created[%d] = %q, want %q", i, got[i], name)
		}
	}

	// The collections with text to compare use a collation.
	for _, name := range []string{model.DeletedUserCollection, model.DisabledUserCollection, model.UserCollection} {
		if len(db.CreateOptions(name)) != 1 {
			t.Errorf("%s: the collation is expected", name)
		}
	}
	if len(db.CreateOptions(model.AuthTokenCollection)) != 0 {
		t.Error("no options are expected for the auth tokens")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
		want bool
	}{
		{"code of the server", mongo.CommandError{Code: 26}, 26, true},
		{"wrapped code of the server", wrap(mongo.CommandError{Code: 27}), 27, true},
		{"other code", mongo.CommandError{Code: 11000, Message: "duplicate"}, 26, false},
		{"message of the driver", errMessage("collection not found with name [a]"), 26, true},
		{"other error", errBoom, 26, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNotFound(tt.err, tt.code, "collection not found with name [a]"); got != tt.want {
				t.Errorf("isNotFound() = %v", got)
			}
		})
	}
}
