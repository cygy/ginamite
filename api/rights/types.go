package rights

import "github.com/gin-gonic/gin"

// Rights : rights of a route (auth token and abilities)
type Rights struct {
	Abilities     []string
	AdminRequired bool
}

// GetUserAbilitesFunc : returns the abilities of a user and a flag to know if he is admin or not.
type GetUserAbilitesFunc func(c *gin.Context, userID string) (bool, []string)
