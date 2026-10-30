// Package robots serves the robots.txt file.
package robots

import (
	"net/http"

	r "github.com/cygy/ginamite/web/robots"

	"github.com/gin-gonic/gin"
)

// Robots sends the content of the robots file.
func Robots(c *gin.Context) {
	c.String(http.StatusOK, r.Content)
}
