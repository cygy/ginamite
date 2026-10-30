package response

import (
	"net/http"
	"strings"
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

func TestResponses_Status(t *testing.T) {
	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	detailed := response.Error{Domain: "domain", Code: 1, Message: "message", Reason: "reason", Recovery: "recovery"}

	tests := []struct {
		name string
		send func(c *gin.Context)
		want int
	}{
		{"Ok", func(c *gin.Context) { Ok(c, H{"a": 1}) }, http.StatusOK},
		{"OkWithStatus", func(c *gin.Context) { OkWithStatus(c, "done") }, http.StatusOK},
		{"Created", func(c *gin.Context) { Created(c, H{"a": 1}) }, http.StatusCreated},
		{"CreatedWithStatus", func(c *gin.Context) { CreatedWithStatus(c, "done") }, http.StatusCreated},
		{"CreatedWithStatusAndID", func(c *gin.Context) { CreatedWithStatusAndID(c, "done", "id") }, http.StatusCreated},
		{"NoContent", NoContent, http.StatusNoContent},
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

func TestOkWithStatus_Body(t *testing.T) {
	c, recorder := apitest.NewContext("GET", "/", nil)
	OkWithStatus(c, "done")

	if body := recorder.Body.String(); !strings.Contains(body, "done") {
		t.Errorf("body = %q, want it to contain the status", body)
	}
}

func TestAddRenewedAuthToken(t *testing.T) {
	c, recorder := apitest.NewContext("GET", "/", nil)
	AddRenewedAuthToken(c, "new-token")

	if got := recorder.Header().Get(response.RenewedAuthTokenHeader); got != "new-token" {
		t.Errorf("header = %q, want new-token", got)
	}
}
