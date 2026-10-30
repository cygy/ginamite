package authentication

import (
	"context"
	"time"
)

// BuildTokenFunc builds a token from its ID.
type BuildTokenFunc func(ctx context.Context, tokenID string) (*Token, error)

// ExtraPropertiesForTokenWithIDFunc returns some extra properties to include in the auth token.
type ExtraPropertiesForTokenWithIDFunc func(tokenID string) (map[string]string, error)

// ExtendTokenExpirationDateFunc extends the expiration date of a token from its ID.
type ExtendTokenExpirationDateFunc func(ctx context.Context, tokenID string, ttl time.Duration) error

var (
	// BuildTokenFromID returns the content of an auth token from its ID.
	BuildTokenFromID BuildTokenFunc

	// ExtraPropertiesForTokenWithID returns some extra properties to include in the auth token.
	ExtraPropertiesForTokenWithID ExtraPropertiesForTokenWithIDFunc

	// ExtendTokenExpirationDateFromID extends the expiration date of an auth token from its ID.
	ExtendTokenExpirationDateFromID ExtendTokenExpirationDateFunc
)
