package functions

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

var errBoom = errors.New("boom")

// withMainSession sets the main session, used by the functions that have no gin context, for the duration of the test.
func withMainSession(t *testing.T, db *mock.Database) {
	t.Helper()

	old := session.Main
	t.Cleanup(func() { session.Main = old })

	session.Main = &session.Session{Store: db}
}

// checkCall checks the method and the first arguments of the last call received by the collection.
func checkCall(t *testing.T, db *mock.Database, collection, method string, args ...any) {
	t.Helper()

	call := db.Coll(collection).LastCall()
	if call.Method != method {
		t.Fatalf("%s: method = %q, want %q (calls: %+v)", collection, call.Method, method, db.Coll(collection).Calls())
	}
	for i, want := range args {
		if i >= len(call.Args) || !reflect.DeepEqual(call.Args[i], want) {
			t.Errorf("%s: argument %d = %#v, want %#v", collection, i, call.Args, want)
		}
	}
}

// checkUpdate checks that the last call of the collection updated the document id with the update.
func checkUpdate(t *testing.T, db *mock.Database, collection string, id bson.ObjectID, update bson.M) {
	t.Helper()

	checkCall(t, db, collection, "UpdateByID", id, update)
}

// calls returns the methods of the calls received by the collection.
func calls(db *mock.Database, collection string) []string {
	var methods []string
	for _, call := range db.Coll(collection).Calls() {
		methods = append(methods, call.Method)
	}

	return methods
}
