package authtoken

import (
	"context"

	jwt "github.com/golang-jwt/jwt/v5"
)

// GetUserAndLocaleFunc returns a struct represnting the user and its loale from an auth token.
type GetUserAndLocaleFunc func(token *jwt.Token) (any, string, error)

// IsValidFunc reports whether the auth token is still valid: neither expired nor revoked. The request fails if it
// returns an error.
type IsValidFunc func(ctx context.Context, tokenID string) (bool, error)
