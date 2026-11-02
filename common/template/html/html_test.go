package html

import (
	"bytes"
	"html/template"
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
	templates := must(t)(LoadTemplates(writeTemplates(t, "<p>{{.Name}}</p>"), nil))

	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "home.html", map[string]string{"Name": "<John>"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "<p>&lt;John&gt;</p>" {
		t.Errorf("got %q", out.String())
	}
}

func TestLoadTemplatesWithDelimiters(t *testing.T) {
	templates := must(t)(LoadTemplatesWithDelimiters(writeTemplates(t, "[[.Name]] {{.Name}}"), "[[", "]]", nil))

	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "home.html", map[string]string{"Name": "John"}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "John {{.Name}}" {
		t.Errorf("got %q", out.String())
	}
}

func TestLoadTemplates_WithEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	must(t)(LoadTemplates(writeTemplates(t, `{{plus 1 2}} {{minus 5 2}} {{mult 3 4}} {{div 8 2}}`), engine))
	engine.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "home.html", nil) })

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK || rec.Body.String() != "3 3 12 4" {
		t.Errorf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestLoadTemplates_HasTheHelpers(t *testing.T) {
	templates := must(t)(LoadTemplates(writeTemplates(t, `{{plus 1 2}}`), nil))

	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, "home.html", nil); err != nil || out.String() != "3" {
		t.Errorf("got %q, %v", out.String(), err)
	}
}

func TestLoadTemplatesWithDelimiters_Engine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	must(t)(LoadTemplatesWithDelimiters(writeTemplates(t, `[[plus 1 2]]`), "[[", "]]", engine))
	engine.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "home.html", nil) })

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Body.String() != "3" {
		t.Errorf("got %q", rec.Body.String())
	}
}

func TestDivisionByZero(t *testing.T) {
	templates := must(t)(LoadTemplates(writeTemplates(t, `{{div 1 0}}`), nil))

	if err := templates.ExecuteTemplate(&bytes.Buffer{}, "home.html", nil); err == nil {
		t.Error("expected an error")
	}
}

func TestLoadTemplates_InvalidTemplate(t *testing.T) {
	engine := gin.New()
	if templates, err := LoadTemplates(writeTemplates(t, "{{.Name"), engine); err == nil || templates != nil {
		t.Errorf("LoadTemplates() = %v, %v, want an error", templates, err)
	}
}

// must returns a function returning the templates, which fails the test on an error.
func must(t *testing.T) func(*template.Template, error) *template.Template {
	t.Helper()

	return func(templates *template.Template, err error) *template.Template {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}

		return templates
	}
}
