package localization

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cygy/ginamite/common/log"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

const frTranslations = `[
  {"id": "hello", "translation": "Bonjour"},
  {"id": "greeting", "translation": "Bonjour {{.Name}}"},
  {"id": "items", "translation": {"one": "{{.Count}} élément", "other": "{{.Count}} éléments"}}
]`

const enTranslations = `[
  {"id": "hello", "translation": "Hello"},
  {"id": "only_en", "translation": "Only in English"}
]`

// writeTranslations writes the translation files, named "<locale>.all.json", and returns their directory.
func writeTranslations(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

// loadEngine returns an engine loaded with a French and an English locale.
func loadEngine(t *testing.T) *Engine {
	t.Helper()

	dir := writeTranslations(t, map[string]string{"fr-fr.all.json": frTranslations, "en-gb.all.json": enTranslations})
	engine := NewEngine()
	engine.Initialize(dir, []string{"fr-FR", "en-GB"})

	return engine
}
