// Package errors defines the errors shared by the application.
// The sub-packages list the error codes returned by the API, grouped by domain.
package errors

import (
	"errors"
	"strings"
)

// NotFoundErrorMessage is the message of the error returned when a resource is not found.
const NotFoundErrorMessage = "not found"

// NotFound returns a new "not found" error.
func NotFound() error {
	return errors.New(NotFoundErrorMessage)
}

// IsNotFound reports whether err is a "not found" error.
// A nil error is not a "not found" error.
func IsNotFound(err error) bool {
	return err != nil && strings.EqualFold(err.Error(), NotFoundErrorMessage)
}
