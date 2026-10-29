package user

import (
	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	"github.com/cygy/ginamite/api/context/shared"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// GetUserID : returns the user ID.
func GetUserID(c *gin.Context) string {
	jwtToken := ctxAuthToken.GetParsedToken(c)
	if jwtToken == nil {
		return ""
	}

	return authentication.GetUserID(jwtToken)
}

// GetUser : returns the user struct.
func GetUser(c *gin.Context) any {
	user, ok := c.Get(shared.UserContextKey)
	if !ok {
		return nil
	}

	return user
}
