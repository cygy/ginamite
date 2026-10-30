package context

import (
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// GetUser : returns a cached user struct from the context.
func GetUser(c *gin.Context) model.User {
	return ctxUser.Get(c).(model.User)
}

// GetFullUser : returns the user from the dabatase.
func GetFullUser(c *gin.Context) *model.User {
	// Get the user from the cache.
	if value, ok := c.Get(fullUserContextKey); ok {
		return value.(*model.User)
	}

	// Get the user from the database.
	mongoSession := session.Get(c)
	userID := ctxUser.GetID(c)

	user := &model.User{}
	if err := user.GetByID(c, userID, mongoSession); err != nil || !user.IsSaved() {
		return nil
	}

	c.Set(fullUserContextKey, user)

	return user
}
