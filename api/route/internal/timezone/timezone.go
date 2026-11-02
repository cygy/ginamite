// Package timezone provides the handlers about the timezones.
package timezone

import (
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/ctxkey"

	"github.com/gin-gonic/gin"
)

// GetTimezones returns the timezones.
func GetTimezones(c *gin.Context) {
	response.Ok(c, response.H{
		"timezones": ctxkey.Timezones.Get(c).Timezones,
	})
}
