// Package user stores and reads the authenticated user in the gin context.
package user

import (
	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	"github.com/cygy/ginamite/api/context/shared"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// GetID returns the user ID.
func GetID(c *gin.Context) string {
	jwtToken := ctxAuthToken.GetParsed(c)
	if jwtToken == nil {
		return ""
	}

	return authentication.GetUserID(jwtToken)
}

// Get returns the user struct.
func Get(c *gin.Context) any {
	user, ok := c.Get(shared.UserContextKey)
	if !ok {
		return nil
	}

	return user
}
