package response

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"
	commonErrors "github.com/cygy/ginamite/common/errors"

	"github.com/gin-gonic/gin"
)

func TestShorthands_Status(t *testing.T) {
	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	tests := []struct {
		name string
		send func(c *gin.Context)
		want int
	}{
		{"UnsupportedRequestLocale", func(c *gin.Context) { UnsupportedRequestLocale(c, "en", "xx", []string{"en"}, "key") }, http.StatusBadRequest},
		{"UnsupportedParameterLocale", func(c *gin.Context) { UnsupportedParameterLocale(c, "en", "xx", "locale", []string{"en"}, "key") }, http.StatusBadRequest},
		{"NotFoundParameterLocale", func(c *gin.Context) { NotFoundParameterLocale(c, "en", "locale", "key") }, http.StatusBadRequest},
		{"UnauthorizedWithInvalidAuthorizationToken", func(c *gin.Context) { UnauthorizedWithInvalidAuthorizationToken(c, "en") }, http.StatusUnauthorized},
		{"UnauthorizedWithRevokedAuthorizationToken", func(c *gin.Context) { UnauthorizedWithRevokedAuthorizationToken(c, "en") }, http.StatusUnauthorized},
		{"UnauthorizedWithExpiredAuthorizationToken", func(c *gin.Context) { UnauthorizedWithExpiredAuthorizationToken(c, "en") }, http.StatusUnauthorized},
		{"UnauthorizedWithInsufficientRights", func(c *gin.Context) { UnauthorizedWithInsufficientRights(c, "en") }, http.StatusUnauthorized},
		{"InvalidRequestParameter", func(c *gin.Context) { InvalidRequestParameter(c, "en") }, http.StatusBadRequest},
		{"InvalidRequestParameterWithDetails", func(c *gin.Context) { InvalidRequestParameterWithDetails(c, "m", "r", "rec") }, http.StatusBadRequest},
		{"InvalidParameterValue", func(c *gin.Context) { InvalidParameterValue(c, "p", "m", "r", "rec") }, http.StatusBadRequest},
		{"NotFoundParameterValue", func(c *gin.Context) { NotFoundParameterValue(c, "p", "m", "r", "rec") }, http.StatusBadRequest},
		{"UnableToReadFile", func(c *gin.Context) { UnableToReadFile(c, "en", "key") }, http.StatusInternalServerError},
		{"UnableToSaveFile", func(c *gin.Context) { UnableToSaveFile(c, "en", "key") }, http.StatusInternalServerError},
		{"FileNotFound", func(c *gin.Context) { FileNotFound(c, "en", "key") }, http.StatusInternalServerError},
		{"MaxFileSizeExceeded", func(c *gin.Context) { MaxFileSizeExceeded(c, "en", "key", 1024) }, http.StatusPreconditionFailed},
		{"InvalidMimeType", func(c *gin.Context) { InvalidMimeType(c, "en", "key") }, http.StatusPreconditionFailed},
		{"UnableToUpdateDatabase", func(c *gin.Context) { UnableToUpdateDatabase(c, "en", "key") }, http.StatusInternalServerError},
		{"SendError not found", func(c *gin.Context) { SendError(c, commonErrors.NotFound()) }, http.StatusNotFound},
		{"SendError other", func(c *gin.Context) { SendError(c, http.ErrAbortHandler) }, http.StatusInternalServerError},
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
