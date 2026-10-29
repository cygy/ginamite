package rights

import "github.com/gin-gonic/gin"

// GetUserAbilitesFunc : returns the abilities of a user and a flag to know if he is admin or not.
type GetUserAbilitesFunc func(c *gin.Context, userID string) (bool, []string)
