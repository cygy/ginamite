package log

import (
	"github.com/cygy/ginamite/common/errors"

	"github.com/sirupsen/logrus"
)

// DatabaseError logs a database error.
func DatabaseError(err error, collection string) {
	if err != nil && !errors.IsNotFound(err) {
		WithFields(logrus.Fields{
			"collection": collection,
			"error":      err.Error(),
		}).Error("database error")
	}
}
