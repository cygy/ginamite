package authtoken

import (
	"time"

	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
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

// RefreshAuthToken : refreshes the auth token with the latest properties of the user.
func RefreshAuthToken(c *gin.Context) {
	tokenID := GetAuthTokenID(c)
	signedToken, err := authentication.RefreshAuthToken(c, config.Main.JWT.Secret, config.Main.JWT.SigningMethod, tokenID, time.Duration(config.Main.JWT.TTL))
	if err == nil {
		response.AddRenewedAuthToken(c, signedToken)
	}
}
