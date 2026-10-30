package context

import "github.com/gin-gonic/gin"

// SavePageKey returns a middleware saving the unique key of the target page to the context.
func SavePageKey(pageKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(pageKeyKey, pageKey)
	}
}

// GetPageKey returns the unique key of the target page saved to the context, or an empty string if there is none.
func GetPageKey(c *gin.Context) string {
	pageKey, _ := c.Get(pageKeyKey)
	value, _ := pageKey.(string)

	return value
}
