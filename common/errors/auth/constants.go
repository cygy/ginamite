// Package auth lists the error codes of the "authentication" domain.
package auth

// Domain is the name of the domain of the errors.
const Domain = "authentication"

// Error codes of the domain.
const (
	// CanNotLogin is the code of the error raised when the login failed.
	CanNotLogin = 0

	// AccountNotFound is the code of the error raised when no account matches the credentials.
	AccountNotFound = 1

	// InvalidAuthorizationToken is the code of the error raised when the authorization token is invalid.
	InvalidAuthorizationToken = 2

	// ExpiredAuthorizationToken is the code of the error raised when the authorization token is expired.
	ExpiredAuthorizationToken = 3

	// RevokedAuthorizationToken is the code of the error raised when the authorization token is revoked.
	RevokedAuthorizationToken = 4

	// InsufficientRights is the code of the error raised when the user has not the rights to do the action.
	InsufficientRights = 5

	// CanNotRefresh is the code of the error raised when the authorization token can not be refreshed.
	CanNotRefresh = 6
)
