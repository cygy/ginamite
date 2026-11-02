// Package context stores and reads the authenticated user of a request, as a model.User, in the gin context.
package context

import (
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

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
