package localization

import "testing"

func TestLoad(t *testing.T) {
	dir := writeTranslations(t, map[string]string{"fr-fr.all.json": frTranslations})

	e, err := Load(dir, []string{"fr-FR"})
	if err != nil {
		t.Fatal(err)
	}

	if got := e.Translate("fr-FR")("hello"); got != "Bonjour" {
		t.Errorf("Translate() = %q", got)
	}
	if e.Translate("en-GB") != nil {
		t.Error("a locale which is not loaded must not have a function")
	}
	if got := e.Source("fr-FR", "greeting"); got != "Bonjour {{.Name}}" {
		t.Errorf("Source() = %q", got)
	}
}

func TestLoad_Error(t *testing.T) {
	if e, err := Load(t.TempDir(), []string{"fr-FR"}); err == nil || e != nil {
		t.Errorf("Load() = %v, %v, want an error", e, err)
	}
}
