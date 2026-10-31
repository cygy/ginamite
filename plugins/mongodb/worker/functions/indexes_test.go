package functions

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"go.mongodb.org/mongo-driver/v2/mongo"
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
