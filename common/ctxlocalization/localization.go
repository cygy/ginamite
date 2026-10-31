// Package ctxlocalization stores and reads the translation engine in the gin context.
package ctxlocalization

import (
	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
)

const key = "_localization"

// Inject returns a middleware saving the translation engine in the gin context, see Get.
func Inject(engine *localization.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		Set(c, engine)
		c.Next()
	}
}

// Set saves the translation engine in the gin context, like Inject does. It is meant for the tests.
func Set(c *gin.Context, engine *localization.Engine) {
	c.Set(key, engine)
}

// Get returns the translation engine saved in the gin context by Inject.
// It panics if there is none: the handlers can not translate their messages without it.
func Get(c *gin.Context) *localization.Engine {
	value, ok := c.Get(key)
	if !ok {
		panic("ctxlocalization: no translation engine in the context, add the middleware Inject")
	}

	return value.(*localization.Engine)
}
