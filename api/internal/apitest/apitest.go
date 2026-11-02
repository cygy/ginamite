// Package apitest provides helpers shared by the tests of the api packages.
package apitest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cygy/ginamite/common/ctxkey"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	log.Initialize("test", false)
}

// Localization is the translation engine of the contexts created by NewContext, set by InitLocalization.
var Localization *localization.Engine

// NewContext returns a gin context bound to a request and the recorder of its response. The context holds the
// translation engine of InitLocalization, if called.
func NewContext(method, target string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, target, body)
	if Localization != nil {
		ctxkey.Localization.Set(c, Localization)
	}

	return c, recorder
}

// Serve sends a request to the handler and returns the recorder of the response.
func Serve(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	return recorder
}

// InitLocalization sets the translation engine, see Localization, with the "en" messages for the duration of the test.
func InitLocalization(t *testing.T, messages map[string]string) {
	t.Helper()

	content, err := json.Marshal(messages)
	if err != nil {
		t.Fatalf("unable to encode the messages: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "en.all.json"), content, 0o600); err != nil {
		t.Fatalf("unable to write the messages: %v", err)
	}

	old := Localization
	engine, err := localization.Load(dir, []string{"en"})
	if err != nil {
		t.Fatal(err)
	}
	Localization = engine
	t.Cleanup(func() { Localization = old })
}
