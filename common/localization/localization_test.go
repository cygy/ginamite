package localization

import "testing"

func TestInitializeAndMainFunctions(t *testing.T) {
	defer func(old *Engine) { Main = old }(Main)

	dir := writeTranslations(t, map[string]string{"fr-fr.all.json": frTranslations})
	Initialize(dir, []string{"fr-FR"})

	if Main == nil {
		t.Fatal("Main is not set")
	}
	if got := Translate("fr-FR")("hello"); got != "Bonjour" {
		t.Errorf("Translate() = %q", got)
	}
	if Translate("en-GB") != nil {
		t.Error("an unknown locale must have no translate function")
	}
	if got := Source("fr-FR", "greeting"); got != "Bonjour {{.Name}}" {
		t.Errorf("Source() = %q", got)
	}
}
