package model

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

var errBoom = errors.New("boom")

// testDocument checks that a document type saves, updates, deletes and loads itself in its collection,
// and keeps its ID.
func testDocument(t *testing.T, collection string, newDocument func() document.Document) {
	t.Helper()

	ctx := context.Background()

	t.Run("ID", func(t *testing.T) {
		doc := newDocument()
		if doc.IsSaved() {
			t.Error("a new document must not be saved")
		}

		id := bson.NewObjectID().Hex()
		if err := doc.SetID(id); err != nil {
			t.Fatal(err)
		}
		if doc.GetID() != id || !doc.IsSaved() {
			t.Errorf("GetID() = %q, IsSaved() = %v", doc.GetID(), doc.IsSaved())
		}

		if err := doc.SetID("nope"); err == nil {
			t.Error("an invalid ID must be rejected")
		}
		if doc.GetID() != id {
			t.Error("a rejected ID must not change the document")
		}
	})

	t.Run("Save", func(t *testing.T) {
		db := mock.NewDatabase("app")
		doc := newDocument()

		id, err := doc.Save(ctx, db)
		if err != nil || id == "" || id != doc.GetID() {
			t.Fatalf("Save() = %q, %v, ID = %q", id, err, doc.GetID())
		}
		if call := db.Coll(collection).LastCall(); call.Method != "InsertOne" || call.Args[0] != doc {
			t.Errorf("call = %+v", call)
		}

		db.Coll(collection).Err = errBoom
		if _, err := newDocument().Save(ctx, db); !errors.Is(err, errBoom) {
			t.Errorf("Save() err = %v", err)
		}
	})

	t.Run("Update", func(t *testing.T) {
		db := mock.NewDatabase("app")
		doc := newDocument()
		_ = doc.SetID(bson.NewObjectID().Hex())

		update := bson.M{"$set": bson.M{"a": 1}}
		if err := doc.Update(ctx, update, db); err != nil {
			t.Fatal(err)
		}

		call := db.Coll(collection).LastCall()
		if call.Method != "UpdateByID" || call.Args[0].(bson.ObjectID).Hex() != doc.GetID() || !reflect.DeepEqual(call.Args[1], update) {
			t.Errorf("call = %+v", call)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		db := mock.NewDatabase("app")
		doc := newDocument()
		_ = doc.SetID(bson.NewObjectID().Hex())

		if err := doc.Delete(ctx, db); err != nil {
			t.Fatal(err)
		}
		if call := db.Coll(collection).LastCall(); call.Method != "DeleteOne" {
			t.Errorf("call = %+v", call)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		db := mock.NewDatabase("app")
		id := bson.NewObjectID()
		db.Coll(collection).One = mock.Result{Value: bson.M{"_id": id}}

		doc := newDocument()
		if err := doc.GetByID(ctx, id.Hex(), db); err != nil {
			t.Fatal(err)
		}
		if doc.GetID() != id.Hex() {
			t.Errorf("ID = %q, want %q", doc.GetID(), id.Hex())
		}

		if err := newDocument().GetByID(ctx, "nope", db); err == nil {
			t.Error("an invalid ID must be rejected")
		}
	})
}

// lastCall returns the last call received by the collection.
func lastCall(db *mock.Database, collection string) mock.Call {
	return db.Coll(collection).LastCall()
}

// checkCall checks the method and the first arguments of the last call received by the collection.
func checkCall(t *testing.T, db *mock.Database, collection, method string, args ...any) {
	t.Helper()

	call := lastCall(db, collection)
	if call.Method != method {
		t.Fatalf("method = %q, want %q (calls: %+v)", call.Method, method, db.Coll(collection).Calls())
	}
	for i, want := range args {
		if i >= len(call.Args) || !reflect.DeepEqual(call.Args[i], want) {
			t.Errorf("argument %d = %#v, want %#v", i, call.Args[i], want)
		}
	}
}

// checkUpdate checks that the last call of the collection updated the document id with the update.
func checkUpdate(t *testing.T, db *mock.Database, collection string, id bson.ObjectID, update bson.M) {
	t.Helper()

	checkCall(t, db, collection, "UpdateByID", id, update)
}

// savedID returns an ObjectID and its hexadecimal form.
func savedID() (bson.ObjectID, string) {
	id := bson.NewObjectID()
	return id, id.Hex()
}

// mustObjectID returns a new ObjectID.
func mustObjectID() bson.ObjectID {
	return bson.NewObjectID()
}

// funcResult is a store.SingleResult whose Decode is a function.
type funcResult func(v any) error

func (f funcResult) Decode(v any) error { return f(v) }
func (f funcResult) Err() error         { return nil }

// equal reports whether the two values are deeply equal.
func equal(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
