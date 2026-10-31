package locale

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"
)

func TestGetLocales(t *testing.T) {
	c, recorder := apitest.NewContext("GET", "/", nil)
	ctxconfig.Set(c, &config.Configuration{SupportedLocales: []string{"en", "fr"}})
	GetLocales(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body struct{ Locales []string }
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid body: %v", err)
	}
	if !slices.Equal(body.Locales, []string{"en", "fr"}) {
		t.Errorf("locales = %v, want [en fr]", body.Locales)
	}
}
