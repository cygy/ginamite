package robots

import (
	"net/http"

	r "github.com/cygy/ginamite/web/robots"

	"github.com/gin-gonic/gin"
)

// Robots : returns the content of the robots file.
func Robots(c *gin.Context) {
	c.String(http.StatusOK, r.Content)
}
