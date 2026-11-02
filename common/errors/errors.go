// Package errors defines the errors shared by the application.
// The sub-packages list the error codes returned by the API, grouped by domain.
package errors

import "errors"

// NotFoundErrorMessage is the message of the error returned when a resource is not found.
const NotFoundErrorMessage = "not found"

// ErrNotFound is the error returned when a resource is not found. Wrap it with %w to add some context.
var ErrNotFound = errors.New(NotFoundErrorMessage)

// IsNotFound reports whether err is, or wraps, ErrNotFound.
// A nil error is not a "not found" error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
