package context

import (
	"time"

	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// GetUserID : returns the user ID from a request.
func GetUserID(c *gin.Context) string {
	token := getJWTToken(c)
	return authentication.GetUserID(token)
}

// GetAuthTokenID : returns the auth token ID from a request.
func GetAuthTokenID(c *gin.Context) string {
	token := getJWTToken(c)
	return authentication.GetTokenID(token)
}

// GetAuthTokenLastUpdate : returns the last update of the auth token from a request.
func GetAuthTokenLastUpdate(c *gin.Context) time.Time {
	token := getJWTToken(c)
	return authentication.GetLastUpdate(token)
}

// IsUserAuthenticated : returns true if the userID is defined into the JWT token.
func IsUserAuthenticated(c *gin.Context) bool {
	userID := GetUserID(c)
	return len(userID) > 0
}
