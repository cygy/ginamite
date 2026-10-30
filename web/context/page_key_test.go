package context

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestSavePageKey(t *testing.T) {
	c, _ := webtest.Context(http.MethodGet, "/", nil)

	if got := GetPageKey(c); got != "" {
		t.Errorf("GetPageKey() = %q before SavePageKey, want empty", got)
	}

	SavePageKey("account.settings")(c)

	if got := GetPageKey(c); got != "account.settings" {
		t.Errorf("GetPageKey() = %q, want account.settings", got)
	}
}

func TestGetPageKey_IgnoresOtherTypes(t *testing.T) {
	c, _ := webtest.Context(http.MethodGet, "/", nil)
	c.Set(pageKeyKey, 42)

	if got := GetPageKey(c); got != "" {
		t.Errorf("GetPageKey() = %q, want empty", got)
	}
}
