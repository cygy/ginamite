package errors

import (
	"errors"
	"strings"
)

// NotFound returns a new error 'not found'.
func NotFound() error {
	return errors.New(NotFoundErrorMessage)
}

// IsNotFound returns true if the error is a 'not found' error.
func IsNotFound(err error) bool {
	return strings.ToLower(err.Error()) == strings.ToLower(NotFoundErrorMessage)
}
