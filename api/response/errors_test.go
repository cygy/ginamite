package response

import (
	"net/http"
	"testing"

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
