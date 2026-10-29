package authtoken

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GetUserAndLocaleFromAuthTokenFunc : returns a struct represnting the user and its loale from an auth token.
type GetUserAndLocaleFromAuthTokenFunc func(token *jwt.Token) (any, string, error)

// IsAuthTokenValidFunc : returns trus if the auth token is still valid and not expired or revoked.
type IsAuthTokenValidFunc func(c *gin.Context, tokenID string) (bool, error)
