package functions

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestCreateIndexes(t *testing.T) {
	fns, db := withMainSession(t)
	index := mongo.IndexModel{Keys: bson.D{{Key: "a", Value: 1}}}

	fns.CreateIndexes(context.Background(), []IndexToCreate{{Collection: "one", Index: index}, {Collection: "two", Index: index}})
	for _, name := range []string{"one", "two"} {
		if view := db.Coll(name).IndexView; view == nil || len(view.Created) != 1 {
			t.Errorf("%s: the index must be created", name)
		}
	}

	// A failure is only logged.
	db.Coll("one").IndexView.Err = errBoom
	fns.CreateIndexes(context.Background(), []IndexToCreate{{Collection: "one", Index: index}})
	if len(db.Coll("one").IndexView.Created) != 2 {
		t.Error("the index must be created again")
	}
}

func TestCreateIndexes_Conflict(t *testing.T) {
	named := mongo.IndexModel{Keys: bson.D{{Key: "a", Value: 1}}, Options: options.Index().SetName("a_1").SetUnique(true)}
	unnamed := mongo.IndexModel{Keys: bson.D{{Key: "a", Value: 1}}, Options: options.Index().SetUnique(true)}
	keySpecs := mongo.CommandError{Code: errorCodeIndexKeySpecsConflict, Message: "An existing index has the same name"}
	optionsConflict := mongo.CommandError{Code: errorCodeIndexOptionsConflict, Message: "Index already exists with different options"}

	tests := []struct {
		name        string
		index       mongo.IndexModel
		createErrs  []error
		err         error
		wantDropped int
		wantCreated int
	}{
		{"key specs conflict replaced", named, []error{keySpecs}, nil, 1, 2},
		{"options conflict replaced", named, []error{optionsConflict}, nil, 1, 2},
		{"index created again with an error", named, []error{keySpecs, errBoom}, nil, 1, 2},
		{"no name to drop", unnamed, []error{keySpecs}, nil, 0, 1},
		{"drop failed", named, []error{keySpecs}, errBoom, 1, 1},
		{"other error", named, []error{errBoom}, nil, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fns, db := withMainSession(t)
			view := &mock.IndexView{CreateErrs: tt.createErrs, Err: tt.err}
			db.Coll("one").IndexView = view

			fns.CreateIndexes(context.Background(), []IndexToCreate{{Collection: "one", Index: tt.index}})
			if len(view.Dropped) != tt.wantDropped || len(view.Created) != tt.wantCreated {
				t.Fatalf("dropped = %v, created %d indexes", view.Dropped, len(view.Created))
			}
			if tt.wantDropped == 1 && view.Dropped[0] != "a_1" {
				t.Errorf("dropped = %v, want a_1", view.Dropped)
			}
		})
	}
}

func TestIndexName(t *testing.T) {
	tests := []struct {
		name  string
		index mongo.IndexModel
		want  string
	}{
		{"no options", mongo.IndexModel{}, ""},
		{"generated name", mongo.IndexModel{Options: options.Index().SetUnique(true)}, ""},
		{"name set", mongo.IndexModel{Options: options.Index().SetUnique(true).SetName("a_1")}, "a_1"},
	}

	for _, tt := range tests {
		if got := indexName(tt.index); got != tt.want {
			t.Errorf("%s: indexName() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestRemoveIndexes(t *testing.T) {
	fns, db := withMainSession(t)

	fns.RemoveIndexes(context.Background(), []IndexToDelete{{Collection: "one", IndexName: "a_1"}, {Collection: "two", IndexName: "b_1"}})
	if got := db.Coll("one").IndexView.Dropped; len(got) != 1 || got[0] != "a_1" {
		t.Errorf("dropped = %v", got)
	}

	// An index that does not exist is not an error, the others are only logged.
	db.Coll("one").IndexView.Err = mongo.CommandError{Code: 27, Message: "index not found"}
	db.Coll("two").IndexView.Err = errBoom
	fns.RemoveIndexes(context.Background(), []IndexToDelete{{Collection: "one", IndexName: "a_1"}, {Collection: "two", IndexName: "b_1"}})
	if len(db.Coll("one").IndexView.Dropped) != 2 || len(db.Coll("two").IndexView.Dropped) != 2 {
		t.Error("every index must be processed")
	}
}

func TestCreateAllIndexes(t *testing.T) {
	fns, db := withMainSession(t)
	fns.CreateAllIndexes(context.Background())

	// Each collection of the plugin has indexes, and every index has keys.
	for _, name := range []string{model.AuthTokenCollection, model.UserCollection, model.DisabledUserCollection, model.UpdatePropertyProcessCollection, model.ForgotPasswordProcessCollection, model.IPLocationCollection, model.ProcessedMessageCollection} {
		view := db.Coll(name).IndexView
		if view == nil || len(view.Created) == 0 {
			t.Errorf("%s: no index created", name)
			continue
		}

		for _, index := range view.Created {
			if index.Keys == nil {
				t.Errorf("%s: an index has no keys", name)
			}
		}
	}
}

func TestRemoveUnusedIndexes(t *testing.T) {
	fns, db := withMainSession(t)
	fns.RemoveUnusedIndexes(context.Background())

	if db.Coll(model.UserCollection).IndexView == nil || len(db.Coll(model.UserCollection).IndexView.Dropped) == 0 {
		t.Error("the unused indexes of the users must be dropped")
	}
	for _, index := range db.Coll(model.UserCollection).IndexView.Dropped {
		if index == "" {
			t.Error("an index has no name")
		}
	}
}
