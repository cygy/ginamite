package apitest

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestSet(t *testing.T) {
	client := Set(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusAccepted) })

	c, _ := webtest.Context(http.MethodGet, "/", nil)
	if result, ok := client.Call(c, api.NewRequest()); !ok || result.StatusCode != http.StatusAccepted {
		t.Errorf("result = %+v, ok = %v", result, ok)
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("http://127.0.0.1", 1)

	c, _ := webtest.Context(http.MethodGet, "/", nil)
	if result, _ := client.Call(c, api.NewRequest()); result.RequestError == nil {
		t.Error("nothing listens on the port 1: the call must fail")
	}
}
