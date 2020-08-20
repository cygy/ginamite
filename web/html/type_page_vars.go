package html

import (
	"fmt"
	"strings"
)

// SetVar : sets up a variable to the list of the variables of the page.
func (page *Page) SetVar(name string, value interface{}) {
	page.initializeVars()
	page.Content.Vars[name] = value
}

// SetLocalizedVar : sets up a localized variable to the list of the variables of the page.
func (page *Page) SetLocalizedVar(name, value string) {
	page.initializeVars()
	page.Content.Vars[name] = page.Localize(value)
}

// SetLocaleVar : sets up a variable to the list of the variables of the page with the locale.
func (page *Page) SetLocaleVar(name, locale string, value interface{}) {
	page.initializeVars()
	key := fmt.Sprintf("%s/%s", name, strings.ToLower(locale))
	page.Content.Vars[key] = value
}

// Var : returns a variable from the list of the variables of the page.
func (page *Page) Var(name string) interface{} {
	return page.Content.Vars[name]
}

// LocaleVar : returns a variable from the list of the variables of the page.
func (page *Page) LocaleVar(name, locale string) interface{} {
	key := fmt.Sprintf("%s/%s", name, strings.ToLower(locale))
	return page.Content.Vars[key]
}

// SetHrefLangParam :
func (page *Page) SetHrefLangParam(locale string, params map[string]string) {
	for i, hrefLang := range page.Meta.HrefLang {
		if hrefLang.Locale == locale {
			page.Meta.HrefLang[i].Href = page.replaceURLVars(hrefLang.Href, params)
			break
		}
	}
}

// Annex functions.
func (page *Page) initializeVars() {
	if page.Content.Vars == nil {
		page.Content.Vars = map[string]interface{}{}
	}
}
