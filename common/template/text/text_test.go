package text

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"text/template"
)

func writeTemplates(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "mails"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mails", "hello.txt"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return dir
}

func render(t *testing.T, templates *template.Template, name string) string {
	t.Helper()

	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, name, map[string]string{"Name": "John"}); err != nil {
		t.Fatal(err)
	}

	return out.String()
}

func TestLoadTemplates(t *testing.T) {
	templates := must(t)(LoadTemplates(writeTemplates(t, "Hello {{.Name}}")))

	if got := render(t, templates, "hello.txt"); got != "Hello John" {
		t.Errorf("got %q", got)
	}
}

func TestLoadTemplatesWithDelimiters(t *testing.T) {
	templates := must(t)(LoadTemplatesWithDelimiters(writeTemplates(t, "Hello [[.Name]] {{.Name}}"), "[[", "]]"))

	if got := render(t, templates, "hello.txt"); got != "Hello John {{.Name}}" {
		t.Errorf("got %q", got)
	}
}

func TestLoadTemplates_InvalidTemplate(t *testing.T) {
	if templates, err := LoadTemplates(writeTemplates(t, "{{.Name")); err == nil || templates != nil {
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
