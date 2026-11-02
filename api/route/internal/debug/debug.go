// Package debug provides the handlers used to debug the API, such as an expired token.
package debug

import (
	"time"

	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/ctxkey"

	"github.com/gin-gonic/gin"
)

// expiredTokenID is the ID of the auth token returned by GetExpiredToken. It matches no stored token.
const expiredTokenID = "expired-token"

// GetExpiredToken returns an auth token, signed by the API, which expired: the clients test their handling of the
// expired tokens with it.
func GetExpiredToken(c *gin.Context) {
	token := &authentication.Token{
		Source:         "debug/test",
		UserID:         "ac006c77fd78cff13e0fa9d1",
		ExpirationDate: time.Now().Add(-1000 * time.Second),
		CreationDate:   time.Now().Add(-10000 * time.Second),
	}

	jwt := ctxkey.Config.Get(c).JWT
	signedToken, err := authentication.SignToken(jwt.SigningMethod, jwt.Secret, expiredTokenID, token, nil)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"auth_token": signedToken,
	})
}
