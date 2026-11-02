// Package request provides helpers about the incoming HTTP requests.
package request

import "github.com/gin-gonic/gin"

const (
	// UserIPAddressHeaderName is the name of the header containing the IP address of the user when the call comes from
	// the web servers. It is read only on the requests of the trusted proxies, see ConfigureClientIP.
	UserIPAddressHeaderName = "X-User-IP"

	// LocaleHeaderName is the name of the HTTP header containing the localization used.
	LocaleHeaderName = "X-Api-Locale"
)

// remoteIPHeaders are the headers giving the address of the client, in the order they are read.
var remoteIPHeaders = []string{UserIPAddressHeaderName, "X-Forwarded-For", "X-Real-IP"}

// ConfigureClientIP makes c.ClientIP() return the address of the client. The headers UserIPAddressHeaderName,
// X-Forwarded-For and X-Real-IP are read only on the requests coming from the trusted proxies, such as a reverse proxy
// or the web servers calling the API: on the other requests the address is the one of the connection, so a client can
// not forge it. No proxy is trusted if trustedProxies is empty. It returns an error if a proxy is not an address or a
// network.
func ConfigureClientIP(engine *gin.Engine, trustedProxies []string) error {
	engine.RemoteIPHeaders = remoteIPHeaders

	return engine.SetTrustedProxies(trustedProxies)
}
