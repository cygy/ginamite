package context

import (
	"time"

	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

// RefreshAuthToken : refreshes the auth token with the latest properties of the user.
func RefreshAuthToken(c *gin.Context) {
	tokenID := GetAuthTokenID(c)

	signedToken, err := authentication.RefreshAuthToken(config.Main.JWT.Secret, config.Main.JWT.SigningMethod, tokenID, time.Duration(config.Main.JWT.TTL))
	if err == nil {
		response.AddRenewedAuthToken(c, signedToken)
	}
}
