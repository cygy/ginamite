package authentication

import "time"

// BuildTokenFunc : function to build a token from its ID.
type BuildTokenFunc func(tokenID string) (*Token, error)

// ExtraPropertiesForTokenWithIDFunc : function to get some extra properties to include in the auth tokens.
type ExtraPropertiesForTokenWithIDFunc func(tokenID string) (map[string]string, error)

// ExtendTokenExpirationDateFunc : function to extend the expiration date of a token from its ID.
type ExtendTokenExpirationDateFunc func(tokenID string, ttl time.Duration) error
