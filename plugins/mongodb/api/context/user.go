// Package context stores and reads the authenticated user of a request, as a model.User, in the gin context.
package context

import (
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"

	"github.com/gin-gonic/gin"
)

// fullUserContextKey is the key of the user loaded from the database in the gin context.
const fullUserContextKey = "_fullUser"

// GetUser returns the user built from the auth token, which holds only a few properties of the user.
// It returns an empty user if the request is not authenticated.
func GetUser(c *gin.Context) model.User {
	switch user := ctxUser.Get(c).(type) {
	case model.User:
		return user
	case *model.User:
		if user != nil {
			return *user
		}
	}

	return model.User{}
}

// GetFullUser returns the authenticated user loaded from the database, or nil if there is none.
// The user is loaded once per request.
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
