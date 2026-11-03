package localization

import "testing"

func TestNewEngine(t *testing.T) {
	e := NewEngine()
	if e.TranslateFuncs == nil || e.Translations == nil {
		t.Errorf("NewEngine() = %+v", e)
	}
}

func TestEngine_Initialize(t *testing.T) {
	e := loadEngine(t)

	if len(e.TranslateFuncs) != 2 || e.TranslateFuncs["fr-FR"] == nil || e.TranslateFuncs["en-GB"] == nil {
		t.Errorf("TranslateFuncs = %v", e.TranslateFuncs)
	}
	if e.Translations["fr-FR"]["hello"] != "Bonjour" {
		t.Errorf("Translations = %v", e.Translations)
	}
}

func TestEngine_Initialize_Errors(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
	}{
		{"missing file", map[string]string{}},
		{"invalid json", map[string]string{"fr-fr.all.json": "["}},
		{"no message", map[string]string{"fr-fr.all.json": "[]"}},
		{"invalid toml", map[string]string{"fr-fr.all.toml": "hello = "}},
		{"empty toml", map[string]string{"fr-fr.all.toml": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeTranslations(t, tt.files)

			if err := NewEngine().Initialize(dir, []string{"fr-FR"}); err == nil {
				t.Error("Initialize() must return an error")
			}
		})
	}
}

func TestEngine_Initialize_TOML(t *testing.T) {
	dir := writeTranslations(t, map[string]string{
		"fr-fr.all.toml": frTOML,
		"en-gb.all.json": enTranslations,
	})

	e := NewEngine()
	if err := e.Initialize(dir, []string{"fr-FR", "en-GB"}); err != nil {
		t.Fatal(err)
	}

	if got := e.Translations["fr-FR"]["error.request.invalid"]; got != "Requête invalide" {
		t.Errorf("dotted key = %q", got)
	}
	if got := e.Translations["fr-FR"]["greeting"]; got != "Bonjour {{.Name}}" {
		t.Errorf("greeting = %q", got)
	}
	if got := e.Translations["en-GB"]["hello"]; got != "Hello" {
		t.Errorf("the JSON locale must still be loaded, hello = %q", got)
	}
}

func TestEngine_Initialize_TOMLPreferred(t *testing.T) {
	dir := writeTranslations(t, map[string]string{
		"fr-fr.all.toml": "hello = \"Salut\"\n",
		"fr-fr.all.json": frTranslations,
	})

	e := NewEngine()
	if err := e.Initialize(dir, []string{"fr-FR"}); err != nil {
		t.Fatal(err)
	}

	if got := e.Translations["fr-FR"]["hello"]; got != "Salut" {
		t.Errorf("hello = %q, the TOML file must win over the JSON file", got)
	}
}

func TestEngine_Translate(t *testing.T) {
	e := loadEngine(t)

	if e.Translate("fr-FR") == nil {
		t.Error("a loaded locale must have a translate function")
	}
	if e.Translate("de-DE") != nil {
		t.Error("an unknown locale must have no translate function")
	}
}

func TestEngine_Source(t *testing.T) {
	e := loadEngine(t)

	tests := []struct {
		locale, key, want string
	}{
		{"fr-FR", "greeting", "Bonjour {{.Name}}"},
		{"fr-FR", "unknown", ""},
		{"de-DE", "hello", ""},
	}

	for _, tt := range tests {
		t.Run(tt.locale+"/"+tt.key, func(t *testing.T) {
			if got := e.Source(tt.locale, tt.key); got != tt.want {
				t.Errorf("Source() = %q, want %q", got, tt.want)
			}
		})
	}
}
