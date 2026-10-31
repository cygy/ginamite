// Package ctxqueue stores and reads the queue sending the messages to the other servers in the gin context.
package ctxqueue

import (
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

const key = "_queue"

// Inject returns a middleware saving the queue in the gin context, see Get.
func Inject(q *queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		Set(c, q)
		c.Next()
	}
}

// Set saves the queue in the gin context, like Inject does. It is meant for the tests.
func Set(c *gin.Context, q *queue.Queue) {
	c.Set(key, q)
}

// Get returns the queue saved in the gin context by Inject.
// It panics if there is none: the handlers can not send messages without it.
func Get(c *gin.Context) *queue.Queue {
	value, ok := c.Get(key)
	if !ok {
		panic("ctxqueue: no queue in the context, add the middleware Inject")
	}

	return value.(*queue.Queue)
}
