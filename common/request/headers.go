package request

import "github.com/gin-gonic/gin"

// hstsValue is the value of the header Strict-Transport-Security: the browsers use HTTPS for one year. The subdomains
// are left out, they may not serve HTTPS.
const hstsValue = "max-age=31536000"

// SecurityHeaders returns a middleware adding the security headers to every response:
//   - X-Content-Type-Options: nosniff, so the browsers do not guess the type of the content;
//   - X-Frame-Options: DENY and Content-Security-Policy: frame-ancestors 'none', so no other site shows the pages in a
//     frame (clickjacking);
//   - Referrer-Policy: strict-origin-when-cross-origin, so the paths and the queries are not sent to the other sites;
//   - Strict-Transport-Security, if hsts is true, so the browsers use HTTPS only. The browsers ignore it over HTTP.
//
// A handler can replace a header with c.Header.
func SecurityHeaders(hsts bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "frame-ancestors 'none'")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if hsts {
			h.Set("Strict-Transport-Security", hstsValue)
		}
	}
}
