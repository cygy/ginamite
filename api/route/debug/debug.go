// Package debug provides the handlers used to debug the API, such as an expired token.
package debug

import (
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/ctxconfig"

	"github.com/gin-gonic/gin"
)

// GetExpiredToken returns an expired token.
func GetExpiredToken(c *gin.Context) {
	signedToken, err := authentication.GetSignedToken(c, ctxconfig.Get(c).JWT.SigningMethod, ctxconfig.Get(c).JWT.Secret, "fakeexpiredtoken")
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"auth_token": signedToken,
	})
}
