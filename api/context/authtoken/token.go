// Package authtoken stores and reads the auth token of a request in the gin context, and provides the middlewares to parse and verify it.
package authtoken

import (
	"time"

	"github.com/cygy/ginamite/common/ctxconfig"

	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GetSigned returns the signed token.
func GetSigned(c *gin.Context) string {
	if value, ok := c.Get(authTokenContextKey); ok {
		return value.(string)
	}

	return ""
}

// GetParsed returns the JWT token.
func GetParsed(c *gin.Context) *jwt.Token {
	if value, ok := c.Get(authTokenParsedContextKey); ok {
		return value.(*jwt.Token)
	}

	return nil
}

// GetID returns the auth token ID.
func GetID(c *gin.Context) string {
	jwtToken := GetParsed(c)
	if jwtToken == nil {
		return ""
	}

	return authentication.GetTokenID(jwtToken)
}

// Refresh refreshes the auth token with the latest properties of the user.
func Refresh(c *gin.Context) {
	tokenID := GetID(c)
	signedToken, err := authentication.RefreshAuthToken(c, ctxconfig.Get(c).JWT.Secret, ctxconfig.Get(c).JWT.SigningMethod, tokenID, time.Duration(ctxconfig.Get(c).JWT.TTL))
	if err == nil {
		response.AddRenewedAuthToken(c, signedToken)
	}
}
