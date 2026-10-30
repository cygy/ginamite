package html

import (
	"testing"

	"github.com/cygy/ginamite/web/internal/webtest"
)

func TestPageVar(t *testing.T) {
	key := "a"
	keyword := "keywordA"

	page := NewPage()

	savedKeyword := page.Var(key)
	if savedKeyword != nil {
		t.Error("A non-saved keyword can not be retrieved.")
	}

	page.SetVar(key, keyword)
	savedKeyword = page.Var(key)

	if savedKeyword.(string) != keyword {
		t.Error("To save a keyword must not modify it.")
	}
}

func TestPageLocaleVar(t *testing.T) {
	key := "a"
	keyword := "keywordA"
	locale := "fr-fr"

	page := NewPage()

	savedKeyword := page.LocaleVar(key, locale)
	if savedKeyword != nil {
		t.Error("A non-saved locale keyword can not be retrieved.")
	}

	page.SetLocaleVar(key, locale, keyword)
	savedKeyword = page.LocaleVar(key, locale)

	if savedKeyword.(string) != keyword {
		t.Error("To save a keyword must not modify it.")
	}

	savedKeyword = page.LocaleVar(key, "en-us")
	if savedKeyword != nil {
		t.Error("A non-saved locale keyword can not be retrieved.")
	}
}

func TestPageLocaleVar_IgnoresTheCaseOfTheLocale(t *testing.T) {
	page := &Page{}

	page.SetLocaleVar("a", "FR-fr", "value")

	if got := page.LocaleVar("a", "fr-FR"); got != "value" {
		t.Errorf("LocaleVar() = %v, want value", got)
	}
	if got := page.Var("a/fr-fr"); got != "value" {
		t.Errorf("Var() = %v, want value", got)
	}
}

func TestPageSetLocalizedVar(t *testing.T) {
	setup(t)

	page := &Page{Locale: "fr-fr"}

	page.SetLocalizedVar("title", "home.title")

	if got := page.Var("title"); got != "Accueil" {
		t.Errorf("Var() = %v, want Accueil", got)
	}
}

func TestPageSetHrefLangParam(t *testing.T) {
	page := &Page{Meta: Meta{HrefLang: []HrefLang{
		{Href: "https://web.test/en-us/account/:id", Locale: "en-us"},
		{Href: "https://web.test/fr-fr/compte/:id", Locale: "fr-fr"},
	}}}

	page.SetHrefLangParam("fr-fr", map[string]string{"id": "42"})
	page.SetHrefLangParam("de-de", map[string]string{"id": "42"})

	if got := page.Meta.HrefLang[1].Href; got != "https://web.test/fr-fr/compte/42" {
		t.Errorf("Href of fr-fr = %q", got)
	}
	if got := page.Meta.HrefLang[0].Href; got != "https://web.test/en-us/account/:id" {
		t.Errorf("Href of en-us = %q, want it unchanged", got)
	}
}

func TestPageRawString(t *testing.T) {
	setup(t)

	page := &Page{Locale: "fr-fr"}

	if got := page.rawString("hello"); got != "Bonjour {{.Name}}" {
		t.Errorf("rawString() = %q, want the untranslated message", got)
	}
	if got := page.rawString("missing"); got != "missing" {
		t.Errorf("rawString() of an unknown key = %q, want the key", got)
	}
}

func TestReplaceVars(t *testing.T) {
	webtest.Hook.Reset()

	tests := []struct {
		name    string
		entry   string
		vars    any
		want    string
		wantLog string
	}{
		{"variable", "Hi {{.Name}}", map[string]string{"Name": "Ann"}, "Hi Ann", ""},
		{"no variable", "plain", nil, "plain", ""},
		{"parse error", "{{.Name", nil, "", "unable to parse the string"},
		{"execute error", "{{.Name.First}}", map[string]string{"Name": "Ann"}, "", "unable to execute the template"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webtest.Hook.Reset()

			if got := replaceVars(tt.entry, tt.vars); got != tt.want {
				t.Errorf("replaceVars() = %q, want %q", got, tt.want)
			}

			entry := webtest.Hook.LastEntry()
			switch {
			case tt.wantLog == "" && entry != nil:
				t.Errorf("unexpected log: %+v", entry)
			case tt.wantLog != "" && (entry == nil || entry.Message != tt.wantLog):
				t.Errorf("logged entry = %+v, want %q", entry, tt.wantLog)
			case entry != nil:
				if _, leaked := entry.Data["vars"]; leaked {
					t.Error("the variables must not be logged")
				}
			}
		})
	}
}

func TestPageReplaceURLVars(t *testing.T) {
	page := &Page{}

	tests := []struct {
		name string
		url  string
		vars []map[string]string
		want string
	}{
		{"no vars", "/a/:id", nil, "/a/:id"},
		{"one var", "/a/:id", []map[string]string{{"id": "1"}}, "/a/1"},
		{"only the first map is used", "/a/:id", []map[string]string{{"id": "1"}, {"id": "2"}}, "/a/1"},
		{"only the first occurrence", "/:id/:id", []map[string]string{{"id": "1"}}, "/1/:id"},
		{"broken template", "{{.", nil, ""},
	}
	for _, tt := range tests {
		if got := page.replaceURLVars(tt.url, tt.vars...); got != tt.want {
			t.Errorf("%s: replaceURLVars() = %q, want %q", tt.name, got, tt.want)
		}
	}
}
