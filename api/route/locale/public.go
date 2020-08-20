package locale

import (
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/config"

	"github.com/gin-gonic/gin"
)

// GetLocales returns the current terms and version.
func GetLocales(c *gin.Context) {
	response.Ok(c, response.H{
		"locales": config.Main.SupportedLocales,
	})
}
