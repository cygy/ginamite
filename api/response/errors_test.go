package response

import (
	"net/http"
	"testing"
	"time"

	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

func TestErrors_Status(t *testing.T) {
	detailed := response.Error{Domain: "domain", Code: 1, Message: "message", Reason: "reason", Recovery: "recovery"}

	tests := []struct {
		name string
		send func(c *gin.Context)
		want int
	}{
		{"NotFound", NotFound, http.StatusNotFound},
		{"NotFoundWithError", func(c *gin.Context) { NotFoundWithError(c, detailed) }, http.StatusNotFound},
		{"InternalServerError", InternalServerError, http.StatusInternalServerError},
		{"InternalServerErrorWithError", func(c *gin.Context) { InternalServerErrorWithError(c, detailed) }, http.StatusInternalServerError},
		{"BadRequest", BadRequest, http.StatusBadRequest},
		{"BadRequestWithError", func(c *gin.Context) { BadRequestWithError(c, detailed) }, http.StatusBadRequest},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized},
		{"UnauthorizedWithError", func(c *gin.Context) { UnauthorizedWithError(c, detailed) }, http.StatusUnauthorized},
		{"Forbidden", Forbidden, http.StatusForbidden},
		{"ForbiddenWithError", func(c *gin.Context) { ForbiddenWithError(c, detailed) }, http.StatusForbidden},
		{"PreconditionFailed", PreconditionFailed, http.StatusPreconditionFailed},
		{"PreconditionFailedWithError", func(c *gin.Context) { PreconditionFailedWithError(c, detailed) }, http.StatusPreconditionFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := apitest.NewContext("GET", "/", nil)
			tt.send(c)
			c.Writer.WriteHeaderNow()

			if recorder.Code != tt.want {
				t.Errorf("status = %d, want %d", recorder.Code, tt.want)
			}
		})
	}
}

func TestTooManyRequests(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter time.Duration
		want       string
	}{
		{"whole seconds", 30 * time.Second, "30"},
		{"rounded up", 29500 * time.Millisecond, "30"},
		{"under a second", 500 * time.Millisecond, "1"},
		{"zero", 0, "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := apitest.NewContext("GET", "/", nil)
			TooManyRequests(c, tt.retryAfter)
			c.Writer.WriteHeaderNow()

			if recorder.Code != http.StatusTooManyRequests {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
			}
			if got := recorder.Header().Get("Retry-After"); got != tt.want {
				t.Errorf("Retry-After = %q, want %q", got, tt.want)
			}
		})
	}
}
