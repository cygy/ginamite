package log

import (
	"github.com/cygy/ginamite/common/errors"

	"github.com/sirupsen/logrus"
)

// DatabaseError logs a database error with the collection it occurred on.
// A nil or "not found" error is not logged.
func DatabaseError(err error, collection string) {
	if err != nil && !errors.IsNotFound(err) {
		WithFields(logrus.Fields{
			"collection": collection,
			"error":      err.Error(),
		}).Error("database error")
	}
}
