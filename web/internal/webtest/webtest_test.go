package webtest

import (
	"net/http"
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
)

func TestHook(t *testing.T) {
	Hook.Reset()

	log.Warn("collected")

	if entry := Hook.LastEntry(); entry == nil || entry.Message != "collected" {
		t.Errorf("entry = %+v", entry)
	}
}

func TestContext(t *testing.T) {
	c, rec := Context(http.MethodPost, "/path?a=1", strings.NewReader("body"))

	c.String(http.StatusAccepted, "ok")

	if c.Request.Method != http.MethodPost || c.Query("a") != "1" || rec.Code != http.StatusAccepted {
		t.Errorf("method = %q, query = %q, status = %d", c.Request.Method, c.Query("a"), rec.Code)
	}
}

func TestServe(t *testing.T) {
	var paths []string
	record := func(c *gin.Context) { paths = append(paths, c.Request.Method+" "+c.Request.URL.Path) }

	rec := Serve(httpRequest(http.MethodDelete, "/a/b"), record, func(c *gin.Context) { c.String(http.StatusTeapot, "tea") })

	if rec.Code != http.StatusTeapot || rec.Body.String() != "tea" || len(paths) != 1 || paths[0] != "DELETE /a/b" {
		t.Errorf("status = %d, body = %q, paths = %v", rec.Code, rec.Body.String(), paths)
	}
}

func TestSetConfig(t *testing.T) {
	original := Config

	t.Run("replaced", func(t *testing.T) {
		SetConfig(t, &config.Configuration{Environment: "test-env"})

		if Config.Environment != "test-env" {
			t.Errorf("Environment = %q", Config.Environment)
		}

		c, _ := Context(http.MethodGet, "/", nil)
		if got := ctxconfig.Get(c).Environment; got != "test-env" {
			t.Errorf("Environment in the context = %q", got)
		}

		var inRequest string
		Serve(httpRequest(http.MethodGet, "/"), func(c *gin.Context) { inRequest = ctxconfig.Get(c).Environment })
		if inRequest != "test-env" {
			t.Errorf("Environment in the request = %q", inRequest)
		}
	})

	if Config != original {
		t.Error("the configuration is not restored")
	}
}

func TestSetTranslations(t *testing.T) {
	original := localization.Main

	t.Run("replaced", func(t *testing.T) {
		SetTranslations(t, map[string]map[string]string{"fr-FR": {"hello": "Bonjour"}, "en-us": {"hello": "Hello"}})

		if got := localization.Translate("fr-FR")("hello"); got != "Bonjour" {
			t.Errorf("translation = %q", got)
		}
		if got := localization.Translate("en-us")("hello"); got != "Hello" {
			t.Errorf("translation = %q", got)
		}
	})

	if localization.Main != original {
		t.Error("the translations are not restored")
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()

	path := WriteFile(t, dir, "a/b/file.txt", "content")

	if content, err := os.ReadFile(path); err != nil || string(content) != "content" || path != filepath.Join(dir, "a/b/file.txt") {
		t.Errorf("path = %q, content = %q, error = %v", path, content, err)
	}
}

func TestSignToken(t *testing.T) {
	signed := SignToken(t, "secret", map[string]any{"sub": "user1"})

	token, err := jwt.Parse(signed, func(*jwt.Token) (any, error) { return []byte("secret"), nil })
	if err != nil || token.Claims.(jwt.MapClaims)["sub"] != "user1" {
		t.Errorf("token = %+v, error = %v", token, err)
	}
}

func TestSetTemplates(t *testing.T) {
	original := htmlTemplate.Main
	engine := gin.New()

	t.Run("replaced", func(t *testing.T) {
		SetTemplates(t, map[string]string{"hello.html": "Hello {{.}}"}, engine)

		var out strings.Builder
		if err := htmlTemplate.Main.ExecuteTemplate(&out, "hello.html", "you"); err != nil || out.String() != "Hello you" {
			t.Errorf("output = %q, error = %v", out.String(), err)
		}
	})

	if htmlTemplate.Main != original {
		t.Error("the templates are not restored")
	}
}

func httpRequest(method, target string) *http.Request {
	c, _ := Context(method, target, nil)

	return c.Request
}
