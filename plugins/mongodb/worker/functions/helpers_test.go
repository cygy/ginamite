package functions

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

var errBoom = errors.New("boom")

// withMainSession sets the main session, used by the functions of the workers, for the duration of the test.
func withMainSession(t *testing.T) *mock.Database {
	t.Helper()

	db := mock.NewDatabase("app")
	old := session.Main
	t.Cleanup(func() { session.Main = old })
	session.Main = &session.Session{Store: db}

	return db
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
