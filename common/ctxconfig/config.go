// Package ctxconfig stores and reads the configuration of the application in the gin context.
package ctxconfig

import (
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

const key = "_config"

// Inject returns a middleware saving the configuration in the gin context, see Get.
func Inject(configuration *config.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		Set(c, configuration)
		c.Next()
	}
}

// Set saves the configuration in the gin context, like Inject does. It is meant for the tests.
func Set(c *gin.Context, configuration *config.Configuration) {
	c.Set(key, configuration)
}

// Get returns the configuration saved in the gin context by Inject.
// It panics if there is none: the handlers can not work without the configuration.
func Get(c *gin.Context) *config.Configuration {
	value, ok := c.Get(key)
	if !ok {
		panic("config: no configuration in the context, add the middleware Inject")
	}

	return value.(*config.Configuration)
}
