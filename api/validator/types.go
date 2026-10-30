// Package validator verifies the tokens issued by the third-party authentication providers.
package validator

import (
	"time"
)

// TokenValidator validates a auth token
type TokenValidator interface {
	SourceName() string
	VerifyTokenAndExtractInfos(token string) (TokenInfos, bool)
}

// TokenInfos struct about the informations of a token
type TokenInfos struct {
	Token     string
	UserID    string
	Email     string
	LastName  string
	FirstName string
	CreatedAt time.Time
	ExpiresAt time.Time
}
