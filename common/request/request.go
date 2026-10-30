// Package request provides helpers about the incoming HTTP requests.
package request

import "github.com/gin-gonic/gin"

const (
	// IPAddressSourceHeaderName is the name of the header containing the IP address of the client behind the reverse proxy.
	IPAddressSourceHeaderName = "X-Real-IP"

	// UserIPAddressHeaderName is the name of the header containing the IP address of the user when the call comes from the web servers.
	UserIPAddressHeaderName = "X-User-IP"

	// LocaleHeaderName is the name of the HTTP header containing the localization used.
	LocaleHeaderName = "X-Api-Locale"
)

// GetRealIPAddress returns the real IP address of the user, even behind a proxy.
// It uses, in this order, the headers UserIPAddressHeaderName and
// IPAddressSourceHeaderName, then the address of the client.
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
