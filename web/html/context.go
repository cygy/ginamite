package html

import "github.com/gin-gonic/gin"

// contextKey is the key of the configuration of the routes in the gin context.
const contextKey = "_pages"

// Inject returns a middleware saving the configuration of the routes in the gin context, see Of.
func Inject(conf *Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		Set(c, conf)
		c.Next()
	}
}

// Set saves the configuration of the routes in the gin context, like Inject does. It is meant for the tests.
func Set(c *gin.Context, conf *Configuration) {
	c.Set(contextKey, conf)
}

// Of returns the configuration of the routes saved in the gin context by Inject.
// It panics if there is none: the handlers can not render the pages without it.
func Of(c *gin.Context) *Configuration {
	value, ok := c.Get(contextKey)
	if !ok {
		panic("html: no configuration of the routes in the context, add the middleware Inject")
	}

	return value.(*Configuration)
}
