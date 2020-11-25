package contact

import (
	"github.com/gin-gonic/gin"
)

func getContactID(c *gin.Context) string {
	return c.Param("id")
}
