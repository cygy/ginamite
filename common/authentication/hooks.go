package authentication

import (
	"context"
	"time"
)

// BuildTokenFunc builds a token from its ID.
type BuildTokenFunc func(ctx context.Context, tokenID string) (*Token, error)

// ExtraPropertiesForTokenWithIDFunc returns some extra properties to include in the auth token. ctx is the context of
// the request signing the token.
type ExtraPropertiesForTokenWithIDFunc func(ctx context.Context, tokenID string) (map[string]string, error)

// ExtendTokenExpirationDateFunc extends the expiration date of a token from its ID.
type ExtendTokenExpirationDateFunc func(ctx context.Context, tokenID string, ttl time.Duration) error

// Hooks are the functions provided by the application, such as the storage plugin, to build the auth tokens and to
// extend their validity.
type Hooks struct {
	// BuildTokenFromID returns the content of an auth token from its ID.
	BuildTokenFromID BuildTokenFunc

	// ExtraPropertiesForTokenWithID returns some extra properties to include in the auth token.
	ExtraPropertiesForTokenWithID ExtraPropertiesForTokenWithIDFunc

	// ExtendTokenExpirationDateFromID extends the expiration date of an auth token from its ID.
	ExtendTokenExpirationDateFromID ExtendTokenExpirationDateFunc
}
