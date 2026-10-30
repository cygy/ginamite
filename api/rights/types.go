package rights

import "github.com/gin-gonic/gin"

// GetUserAbilitiesFunc returns the abilities of a user and a flag to know if he is admin or not.
type GetUserAbilitiesFunc func(c *gin.Context, userID string) (bool, []string)
