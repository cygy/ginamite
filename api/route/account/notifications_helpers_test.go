package account

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

const validNotificationID = "65f0c0ffee65f0c0ffee65f0"

func TestCheckNotificationID(t *testing.T) {
	tests := []struct {
		name   string
		params gin.Params
		want   bool
		status int
	}{
		{"missing", nil, false, http.StatusBadRequest},
		{"not a valid ID", gin.Params{{Key: "id", Value: "nope"}}, false, http.StatusBadRequest},
		{"valid", gin.Params{{Key: "id", Value: validNotificationID}}, true, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)

			c, recorder := newContext(tt.params)
			id, ok := checkNotificationID(c, "placeholder")
			if ok != tt.want {
				t.Fatalf("ok = %v, want %v", ok, tt.want)
			}
			if ok && id != validNotificationID {
				t.Errorf("id = %q, want %q", id, validNotificationID)
			}
			c.Writer.WriteHeaderNow()
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d", recorder.Code, tt.status)
			}
		})
	}
}
