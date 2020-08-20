package html

import "testing"

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
