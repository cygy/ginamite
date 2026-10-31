// Package webtest provides helpers shared by the tests of the web server: a logger collecting its entries, Gin
// contexts, and the replacement of the global configuration and translations.
package webtest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	htmlTemplate "github.com/cygy/ginamite/common/template/html"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

// Hook collects the entries of the main logger, which discards them.
var Hook *logtest.Hook

// Config is the configuration saved in the context of the requests created by Context and Serve, see SetConfig.
var Config *config.Configuration

func init() {
	gin.SetMode(gin.TestMode)

	var logger *logrus.Logger
	logger, Hook = logtest.NewNullLogger()
	log.Main = logger.WithField("server", "test")
}

// Context returns a Gin context for a request, and the recorder of its response. The context holds the configuration
// of SetConfig, if any.
func Context(method, target string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, target, body)
	if Config != nil {
		ctxconfig.Set(c, Config)
	}

	return c, rec
}

// Serve sends a request through the handlers, registered for any method on any path, and returns the response.
// The requests hold the configuration of SetConfig, if any.
func Serve(req *http.Request, handlers ...gin.HandlerFunc) *httptest.ResponseRecorder {
	router := gin.New()
	if Config != nil {
		router.Use(ctxconfig.Inject(Config))
	}
	router.Any("/*path", handlers...)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

// SetConfig replaces the configuration saved in the context of the requests, see Config, for the duration of the test.
func SetConfig(t *testing.T, conf *config.Configuration) {
	t.Helper()

	old := Config
	Config = conf
	t.Cleanup(func() { Config = old })
}

// SetTranslations replaces the main localization engine for the duration of the test, with the messages of each locale.
func SetTranslations(t *testing.T, messages map[string]map[string]string) {
	t.Helper()

	dir := t.TempDir()
	locales := make([]string, 0, len(messages))
	for locale, translations := range messages {
		content, err := json.Marshal(translations)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(dir, strings.ToLower(locale)+".all.json"), content, 0o600); err != nil {
			t.Fatal(err)
		}
		locales = append(locales, locale)
	}

	old := localization.Main
	localization.Initialize(dir, locales)
	t.Cleanup(func() { localization.Main = old })
}

// WriteFile writes a file, and its missing parent directories, in the directory dir and returns its path.
func WriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

// SignToken returns a JWT with the claims, signed with HS256 and the secret.
func SignToken(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims)).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	return signed
}

// SetTemplates writes the templates, by file name, loads them into the main HTML templates and into the Gin engine,
// if not nil, and restores the main HTML templates at the end of the test.
func SetTemplates(t *testing.T, templates map[string]string, e *gin.Engine) {
	t.Helper()

	dir := t.TempDir()
	for name, content := range templates {
		WriteFile(t, dir, "pages/"+name, content)
	}

	old := htmlTemplate.Main
	htmlTemplate.LoadTemplates(dir, e)
	t.Cleanup(func() { htmlTemplate.Main = old })
}
