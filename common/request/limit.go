package request

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LimitBodySize returns a middleware rejecting the bodies larger than maxSize bytes, so that a client can not exhaust
// the memory or the disk of the server. A request announcing a larger body is answered at once with the status 413;
// the reading of a body without length, such as a chunked one, fails with a *http.MaxBytesError past the limit. Nothing
// is limited if maxSize is zero or negative.
func LimitBodySize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxSize <= 0 || c.Request.Body == nil {
			return
		}

		if c.Request.ContentLength > maxSize {
			c.AbortWithStatus(http.StatusRequestEntityTooLarge)
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
	}
}
