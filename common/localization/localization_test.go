package localization

import "testing"

func TestLoad(t *testing.T) {
	dir := writeTranslations(t, map[string]string{"fr-fr.all.json": frTranslations})

	e := Load(dir, []string{"fr-FR"})

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
