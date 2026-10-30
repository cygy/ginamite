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

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	log.Initialize("test", false)
}

// NewContext returns a gin context bound to a request and the recorder of its response.
func NewContext(method, target string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, target, body)

	return c, recorder
}

// Serve sends a request to the handler and returns the recorder of the response.
func Serve(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	return recorder
}

// InitLocalization initializes the translation engine with the "en" messages.
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

	localization.Initialize(dir, []string{"en"})
}
