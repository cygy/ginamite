// Package request lists the error codes of the "request" domain.
package request

// Domain is the name of the domain of the errors.
const Domain = "request"

// InvalidCharactersSet lists the characters forbidden in the value of a JSON field.
const InvalidCharactersSet = "'‘{}"

// Error codes of the domain.
const (
	// InvalidParameterValue is the code of the error raised when a parameter has an invalid value.
	InvalidParameterValue = 0

	// NotFoundParameterValue is the code of the error raised when a mandatory parameter is missing.
	NotFoundParameterValue = 1

	// InvalidRequest is the code of the error raised when the request is invalid.
	InvalidRequest = 2
)
