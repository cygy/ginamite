package timezone

import (
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/gin-gonic/gin"
)

// GetTimezones : returns the timezones.
func GetTimezones(c *gin.Context) {
	response.Ok(c, response.H{
		"timezones": timezone.Main.Timezones,
	})
}
