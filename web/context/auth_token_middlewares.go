package context

import (
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SaveAuthToken : returns a middleware which saves the auth token from the request to the context.
func SaveAuthToken(cookieName, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the signed token.
		signedToken, err := c.Cookie(cookieName)
		if err != nil || len(signedToken) == 0 {
			return
		}

		// Verify the JWT token.
		token, err := authentication.ParseToken(signedToken, secret)
		if token == nil || err != nil {
			log.WithFields(logrus.Fields{
				"token": signedToken,
			}).Error("unable to validate the authentication token from the cookie")
			return
		}

		c.Set(tokenKey, token)
	}
}
