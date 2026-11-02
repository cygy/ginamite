package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHello(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/hello", nil)

	hello(c)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d", recorder.Code)
	}
	if got, want := recorder.Body.String(), `{"hello":"world"}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}
}
