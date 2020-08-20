package debug

import (
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

// GetExpiredToken : returns an expired token.
func GetExpiredToken(c *gin.Context) {
	signedToken, err := authentication.GetSignedToken(config.Main.JWT.SigningMethod, config.Main.JWT.Secret, "fakeexpiredtoken")
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"auth_token": signedToken,
	})
}
