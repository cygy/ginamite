package response

import (
	"github.com/cygy-solutions/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// AddRenewedAuthToken adds a header to the response.
func AddRenewedAuthToken(c *gin.Context, token string) {
	c.Writer.Header().Set(response.RenewedAuthTokenHeader, token)
}
