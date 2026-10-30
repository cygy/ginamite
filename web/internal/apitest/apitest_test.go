package apitest

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestSet(t *testing.T) {
	original := api.Main

	t.Run("replaced", func(t *testing.T) {
		Set(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusAccepted) })

		c, _ := webtest.Context(http.MethodGet, "/", nil)
		if result, ok := api.Main.Call(c, api.NewRequest()); !ok || result.StatusCode != http.StatusAccepted {
			t.Errorf("result = %+v, ok = %v", result, ok)
		}
	})

	if api.Main != original {
		t.Error("the API client is not restored")
	}
}

func TestSetAddress(t *testing.T) {
	original := api.Main

	t.Run("replaced", func(t *testing.T) {
		SetAddress(t, "http://127.0.0.1", 1)

		c, _ := webtest.Context(http.MethodGet, "/", nil)
		if result, _ := api.Main.Call(c, api.NewRequest()); result.RequestError == nil {
			t.Error("nothing listens on the port 1: the call must fail")
		}
	})

	if api.Main != original {
		t.Error("the API client is not restored")
	}
}
