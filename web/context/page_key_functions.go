package context

import (
	"github.com/gin-gonic/gin"
)

// SavePageKey : saves the unique key of the target page to the context.
func SavePageKey(pageKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(pageKeyKey, pageKey)
	}
}

// GetPageKey : returns the unique key of the target page from the context.
func GetPageKey(c *gin.Context) string {
	if value, ok := c.Get(pageKeyKey); ok {
		return value.(string)
	}

	return ""
}
