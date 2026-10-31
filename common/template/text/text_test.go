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
	templates := LoadTemplates(writeTemplates(t, "Hello {{.Name}}"))

	if got := render(t, templates, "hello.txt"); got != "Hello John" {
		t.Errorf("got %q", got)
	}
}

func TestLoadTemplatesWithDelimiters(t *testing.T) {
	templates := LoadTemplatesWithDelimiters(writeTemplates(t, "Hello [[.Name]] {{.Name}}"), "[[", "]]")

	if got := render(t, templates, "hello.txt"); got != "Hello John {{.Name}}" {
		t.Errorf("got %q", got)
	}
}

func TestLoadTemplates_InvalidTemplate(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("LoadTemplates() must panic on an invalid template")
		}
	}()

	LoadTemplates(writeTemplates(t, "{{.Name"))
}
