package weburl

import "github.com/gin-gonic/gin"

// contextKey is the key of the routes in the gin context.
const contextKey = "_weburl"

// Inject returns a middleware saving the routes in the gin context, see Of. The routes may be nil.
func Inject(routes *Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(contextKey, routes)
		c.Next()
	}
}

// Of returns the routes saved in the gin context by Inject, or nil if there is none: the server runs without web
// routes file, or without the middleware. A nil Configuration is safe to use, its URLs are empty.
func Of(c *gin.Context) *Configuration {
	value, _ := c.Get(contextKey)
	routes, _ := value.(*Configuration)

	return routes
}
