package context

import (
	"github.com/cygy/ginamite/common/authentication"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// GetSignedToken : returns the signed token.
func GetSignedToken(c *gin.Context) string {
	if value, ok := c.Get(authTokenContextKey); ok {
		return value.(string)
	}

	return ""
}

// GetParsedToken : returns the JWT token.
func GetParsedToken(c *gin.Context) *jwt.Token {
	if value, ok := c.Get(authTokenParsedContextKey); ok {
		return value.(*jwt.Token)
	}

	return nil
}

// GetAuthTokenID : returns the auth token ID.
func GetAuthTokenID(c *gin.Context) string {
	jwtToken := GetParsedToken(c)

	if jwtToken == nil {
		return ""
	}

	return authentication.GetTokenID(jwtToken)
}

// GetUserID : returns the user ID.
func GetUserID(c *gin.Context) string {
	jwtToken := GetParsedToken(c)

	if jwtToken == nil {
		return ""
	}

	return authentication.GetUserID(jwtToken)
}

// GetUser : returns the user struct.
func GetUser(c *gin.Context) interface{} {
	user, ok := c.Get(userContextKey)
	if !ok {
		return nil
	}

	return user
}
