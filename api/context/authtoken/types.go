package authtoken

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GetUserAndLocaleFunc returns a struct represnting the user and its loale from an auth token.
type GetUserAndLocaleFunc func(token *jwt.Token) (any, string, error)

// IsValidFunc returns trus if the auth token is still valid and not expired or revoked.
type IsValidFunc func(c *gin.Context, tokenID string) (bool, error)
