package status

import (
	"net/http"
	"strings"
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"
)

func TestGetStatus(t *testing.T) {
	c, recorder := apitest.NewContext("GET", "/", nil)
	GetStatus(c)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "ok") {
		t.Errorf("body = %q, want it to contain ok", body)
	}
}
