// Package ctxtimezone stores and reads the timezones of the countries in the gin context.
package ctxtimezone

import (
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

const key = "_timezones"

// Inject returns a middleware saving the timezones in the gin context, see Get.
func Inject(timezones *timezone.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		Set(c, timezones)
		c.Next()
	}
}

// Set saves the timezones in the gin context, like Inject does. It is meant for the tests.
func Set(c *gin.Context, timezones *timezone.Configuration) {
	c.Set(key, timezones)
}

// Get returns the timezones saved in the gin context by Inject.
// It panics if there are none: the handlers can not check the timezones without them.
func Get(c *gin.Context) *timezone.Configuration {
	value, ok := c.Get(key)
	if !ok {
		panic("ctxtimezone: no timezones in the context, add the middleware Inject")
	}

	return value.(*timezone.Configuration)
}
