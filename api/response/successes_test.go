package response

import (
	"net/http"
	"strings"
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

func TestSuccesses_Status(t *testing.T) {
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
