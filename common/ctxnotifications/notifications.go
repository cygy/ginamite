// Package ctxnotifications stores and reads the registry of the notifications in the gin context.
package ctxnotifications

import (
	"github.com/cygy/ginamite/common/notifications"

	"github.com/gin-gonic/gin"
)

const key = "_notifications"

// Inject returns a middleware saving the registry of the notifications in the gin context, see Get.
func Inject(registry *notifications.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		Set(c, registry)
		c.Next()
	}
}

// Set saves the registry in the gin context, like Inject does. It is meant for the tests.
func Set(c *gin.Context, registry *notifications.Registry) {
	c.Set(key, registry)
}

// Get returns the registry saved in the gin context by Inject.
// It panics if there is none: the handlers can not check the notifications without it.
func Get(c *gin.Context) *notifications.Registry {
	value, ok := c.Get(key)
	if !ok {
		panic("ctxnotifications: no registry in the context, add the middleware Inject")
	}

	return value.(*notifications.Registry)
}
