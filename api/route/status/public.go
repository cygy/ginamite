package status

import (
	"github.com/cygy/ginamite/api/response"
	r "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// GetStatus returns the status of the service ("ok")
func GetStatus(c *gin.Context) {
	response.Ok(c, r.NewStatus("ok"))
}
