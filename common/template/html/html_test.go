package html

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func writeTemplates(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "pages"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pages", "home.html"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestLoadTemplates_WithoutEngine(t *testing.T) {
	defer func() { Main = nil }()

	LoadTemplates(writeTemplates(t, "<p>{{.Name}}</p>"), nil)

	var out bytes.Buffer
	if err := Main.ExecuteTemplate(&out, "home.html", map[string]string{"Name": "<John>"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "<p>&lt;John&gt;</p>" {
		t.Errorf("got %q", out.String())
	}
}

func TestLoadTemplatesWithDelimiters(t *testing.T) {
	defer func() { Main = nil }()

	LoadTemplatesWithDelimiters(writeTemplates(t, "[[.Name]] {{.Name}}"), "[[", "]]", nil)

	var out bytes.Buffer
	if err := Main.ExecuteTemplate(&out, "home.html", map[string]string{"Name": "John"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "John {{.Name}}" {
		t.Errorf("got %q", out.String())
	}
}

func TestLoadTemplates_WithEngine(t *testing.T) {
	defer func() { Main = nil }()
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	LoadTemplates(writeTemplates(t, `{{plus 1 2}} {{minus 5 2}} {{mult 3 4}} {{div 8 2}}`), engine)
	engine.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "home.html", nil) })

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK || rec.Body.String() != "3 3 12 4" {
		t.Errorf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestLoadTemplates_MainHasTheHelpers(t *testing.T) {
	defer func() { Main = nil }()

	LoadTemplates(writeTemplates(t, `{{plus 1 2}}`), nil)

	var out bytes.Buffer
	if err := Main.ExecuteTemplate(&out, "home.html", nil); err != nil || out.String() != "3" {
		t.Errorf("got %q, %v", out.String(), err)
	}
}

func TestLoadTemplatesWithDelimiters_Engine(t *testing.T) {
	defer func() { Main = nil }()
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	LoadTemplatesWithDelimiters(writeTemplates(t, `[[plus 1 2]]`), "[[", "]]", engine)
	engine.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "home.html", nil) })

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Body.String() != "3" {
		t.Errorf("got %q", rec.Body.String())
	}
}

func TestDivisionByZero(t *testing.T) {
	defer func() { Main = nil }()

	LoadTemplates(writeTemplates(t, `{{div 1 0}}`), nil)

	if err := Main.ExecuteTemplate(&bytes.Buffer{}, "home.html", nil); err == nil {
		t.Error("expected an error")
	}
}

func TestLoadTemplates_InvalidTemplate(t *testing.T) {
	defer func() { Main = nil }()
	defer func() {
		if recover() == nil {
			t.Error("LoadTemplates() must panic on an invalid template")
		}
	}()

	LoadTemplates(writeTemplates(t, "{{.Name"), nil)
}
