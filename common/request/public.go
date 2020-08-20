package request

import "github.com/gin-gonic/gin"

// GetRealIPAddress : returns the real address IP of the user, if it is behind a proxy.
func GetRealIPAddress(c *gin.Context) string {
	source := c.Request.Header.Get(UserIPAddressHeaderName)

	if len(source) == 0 {
		source = c.Request.Header.Get(IPAddressSourceHeaderName)
	}

	if len(source) == 0 {
		source = c.ClientIP()
	}

	return source
}
