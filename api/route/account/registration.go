package account

import (
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// RegisterWithPassword creates a new account.
func RegisterWithPassword(c *gin.Context) {
	register(c, authentication.MethodPassword)
}

// RegisterWithFacebook creates a new account.
func RegisterWithFacebook(c *gin.Context) {
	register(c, authentication.MethodFacebook)
}

// RegisterWithGoogle creates a new account.
func RegisterWithGoogle(c *gin.Context) {
	register(c, authentication.MethodGoogle)
}

// ValidateRegistration validates a new account.
func ValidateRegistration(c *gin.Context) {
	validateOrCancelRegistration(c, true)
}

// CancelRegistration cancels a new account.
func CancelRegistration(c *gin.Context) {
	validateOrCancelRegistration(c, false)
}
