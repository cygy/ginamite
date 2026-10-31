package timezone

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/cygy/ginamite/common/ctxtimezone"

	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/timezone"
)

func TestGetTimezones(t *testing.T) {
	c, recorder := apitest.NewContext("GET", "/", nil)
	ctxtimezone.Set(c, &timezone.Configuration{Timezones: []string{"Europe/Paris", "UTC"}})
	GetTimezones(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body struct{ Timezones []string }
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid body: %v", err)
	}
	if !slices.Equal(body.Timezones, []string{"Europe/Paris", "UTC"}) {
		t.Errorf("timezones = %v, want [Europe/Paris UTC]", body.Timezones)
	}
}
