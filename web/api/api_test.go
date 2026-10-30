package api

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

func TestInitialize(t *testing.T) {
	old := Main
	t.Cleanup(func() { Main = old })

	Initialize("http://127.0.0.1", "v1", 8080, 1000, 2, false)

	if Main == nil || Main.client == nil {
		t.Fatalf("Main = %+v", Main)
	}
	if got := Main.client.BaseURL; got != "http://127.0.0.1:8080/v1" {
		t.Errorf("base URL = %q", got)
	}
	if Main.client.RetryCount != 2 {
		t.Errorf("retry count = %d", Main.client.RetryCount)
	}
}

func TestAddResponseHandlers(t *testing.T) {
	old := Main
	t.Cleanup(func() { Main = old })
	Main = NewClient("http://127.0.0.1", "v1", 1, 1000, 0, false)

	called := false
	AddResponseHandlers(func(c *gin.Context, statusCode int, requestError error, apiError response.Error) bool {
		called = true
		return true
	})

	if len(Main.responseHandlers) != 1 {
		t.Fatalf("%d handlers, want 1", len(Main.responseHandlers))
	}
	c, _ := webtest.Context(http.MethodGet, "/", nil)
	Main.responseHandlers[0](c, 0, nil, response.Error{})
	if !called {
		t.Error("the handler was not added")
	}
}
