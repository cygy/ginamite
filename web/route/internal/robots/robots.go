// Package robots serves the robots.txt file.
package robots

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Robots returns the handler sending the content of the robots file.
func Robots(content string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, content)
	}
}
