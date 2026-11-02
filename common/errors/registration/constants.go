// Package registration lists the error codes of the "registration" domain.
package registration

// Domain is the name of the domain of the errors.
const Domain = "registration"

// Error codes of the domain.
const (
	// AccountAlreadyExists is the code of the error raised when an account already exists.
	AccountAlreadyExists = 0

	// TokenCheckFailed is the code of the error raised when the Facebook or Google auth token can not be verified, it may have been altered.
	TokenCheckFailed = 1
)
