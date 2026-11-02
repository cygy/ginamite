package request

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// CrossOriginProtection returns a middleware rejecting with the status 403 the cross-origin requests whose method is
// not safe (POST, PUT, PATCH, DELETE...), to protect the forms and the calls authenticated by a cookie from the CSRF
// attacks. A request is cross-origin when the header Sec-Fetch-Site says so, or when the host of its header Origin
// is not the one of the request, see http.CrossOriginProtection. The requests without these headers, which do not
// come from a browser, are allowed. trustedOrigins, such as "https://www.example.com", are allowed too; the empty
// ones are ignored. It returns an error if an origin is not a scheme and a host.
func CrossOriginProtection(trustedOrigins ...string) (gin.HandlerFunc, error) {
	protection := http.NewCrossOriginProtection()

	for _, origin := range trustedOrigins {
		if origin == "" {
			continue
		}

		u, err := url.Parse(origin)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("invalid trusted origin %q", origin)
		}

		if err := protection.AddTrustedOrigin(u.Scheme + "://" + u.Host); err != nil {
			return nil, fmt.Errorf("invalid trusted origin %q: %w", origin, err)
		}
	}

	return func(c *gin.Context) {
		if err := protection.Check(c.Request); err != nil {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}, nil
}
