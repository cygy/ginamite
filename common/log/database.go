package log

import (
	"github.com/cygy/ginamite/common/errors"
)

// DatabaseError logs a database error with the collection it occurred on.
// A nil or "not found" error is not logged.
func DatabaseError(err error, collection string) {
	if err != nil && !errors.IsNotFound(err) {
		Error("database error", "collection", collection, "error", err.Error())
	}
}
