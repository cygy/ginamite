package functions

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

var errBoom = errors.New("boom")

// withMainSession returns the functions of the workers working on a stub database, and the database.
func withMainSession(t *testing.T) (*Functions, *mock.Database) {
	t.Helper()

	db := mock.NewDatabase("app")

	return New(func() store.Database { return db }), db
}

// methods returns the methods of the calls received by the collection.
func methods(db *mock.Database, collection string) []string {
	var names []string
	for _, call := range db.Coll(collection).Calls() {
		names = append(names, call.Method)
	}

	return names
}

// errMessage is an error with a message.
type errMessage string

func (e errMessage) Error() string { return string(e) }

// wrap wraps the error.
func wrap(err error) error {
	return fmt.Errorf("context: %w", err)
}
