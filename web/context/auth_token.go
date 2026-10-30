package context

import (
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// SaveAuthToken returns a middleware saving the auth token found in the cookie cookieName to the context.
// Nothing is saved if the cookie is missing or if its token is invalid, which is logged without the token.
func SaveAuthToken(cookieName, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		signedToken, err := c.Cookie(cookieName)
		if err != nil || len(signedToken) == 0 {
			return
		}

		token, err := authentication.ParseToken(signedToken, secret)
		if token == nil || err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("unable to validate the authentication token from the cookie")
			return
		}

		c.Set(tokenKey, token)
	}
}

// GetUserID returns the ID of the user of the request, or an empty string if there is none.
func GetUserID(c *gin.Context) string {
	return authentication.GetUserID(getJWTToken(c))
}

// GetAuthTokenID returns the ID of the auth token of the request, or an empty string if there is none.
func GetAuthTokenID(c *gin.Context) string {
	return authentication.GetTokenID(getJWTToken(c))
}

// GetAuthTokenLastUpdate returns the date of the last update of the auth token of the request,
// or the Unix epoch if there is none.
func GetAuthTokenLastUpdate(c *gin.Context) time.Time {
	return authentication.GetLastUpdate(getJWTToken(c))
}

// IsUserAuthenticated reports whether the request has a token with a user ID.
func IsUserAuthenticated(c *gin.Context) bool {
	return len(GetUserID(c)) > 0
}

// getJWTToken returns the token saved by SaveAuthToken, or nil if there is none.
func getJWTToken(c *gin.Context) *jwt.Token {
	token, _ := c.Get(tokenKey)
	jwtToken, _ := token.(*jwt.Token)

	return jwtToken
}
