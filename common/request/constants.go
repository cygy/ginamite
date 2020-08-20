package request

const (
	// IPAddressSourceHeaderName : name of the header containing the IP address behind the reverse proxy.
	IPAddressSourceHeaderName = "X-Real-IP"

	// UserIPAddressHeaderName : name of the header containing the IP address of the user if the call is from teh web servers.
	UserIPAddressHeaderName = "X-User-IP"

	// LocaleHeaderName http header name of the localization used.
	LocaleHeaderName = "X-Api-Locale"
)
