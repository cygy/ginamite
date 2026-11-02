package functions

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
)

var errBoom = errors.New("boom")

// tokensOn returns the functions of the auth tokens working on the database.
func tokensOn(db store.Database) *Tokens {
	return NewTokens(func() store.Database { return db })
}

// processesOn returns the functions of the processes working on the database.
func processesOn(db store.Database) *Processes {
	return NewProcesses(func() store.Database { return db })
}

// accountsOn returns the functions of the accounts working on the database.
func accountsOn(db store.Database) *Accounts {
	return NewAccounts(func() store.Database { return db }, model.DefaultSettings,
		func(string) model.RegistrationDefaults { return model.RegistrationDefaults{Timezone: "UTC"} },
		func(string) {})
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
