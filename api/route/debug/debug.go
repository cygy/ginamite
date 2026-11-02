// Package debug provides the handlers used to debug the API, such as an expired token.
package debug

import (
	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/ctxkey"

	"github.com/gin-gonic/gin"
)

// GetExpiredToken returns an expired token.
func GetExpiredToken(c *gin.Context) {
	signedToken, err := ctxAuthToken.HooksOf(c).GetSignedToken(c, ctxkey.Config.Get(c).JWT.SigningMethod, ctxkey.Config.Get(c).JWT.Secret, "fakeexpiredtoken")
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"auth_token": signedToken,
	})
}
