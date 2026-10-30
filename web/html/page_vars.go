package html

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"

	"github.com/sirupsen/logrus"
)

// SetVar defines a variable of the page.
func (page *Page) SetVar(name string, value any) {
	page.initializeVars()
	page.Content.Vars[name] = value
}

// SetLocalizedVar defines a variable of the page, whose value is the translation of the message ID value.
func (page *Page) SetLocalizedVar(name, value string) {
	page.initializeVars()
	page.Content.Vars[name] = page.Localize(value)
}

// SetLocaleVar defines a variable of the page for a locale. See LocaleVar.
func (page *Page) SetLocaleVar(name, locale string, value any) {
	page.initializeVars()
	page.Content.Vars[localeVarKey(name, locale)] = value
}

// Var returns a variable of the page, or nil if it is not defined.
func (page *Page) Var(name string) any {
	return page.Content.Vars[name]
}

// LocaleVar returns the variable name of the page for a locale, whatever its case, or nil if it is not defined.
func (page *Page) LocaleVar(name, locale string) any {
	return page.Content.Vars[localeVarKey(name, locale)]
}

// SetHrefLangParam replaces the parameters, prefixed by a colon, of the alternate URL of the locale.
func (page *Page) SetHrefLangParam(locale string, params map[string]string) {
	for i, hrefLang := range page.Meta.HrefLang {
		if hrefLang.Locale == locale {
			page.Meta.HrefLang[i].Href = page.replaceURLVars(hrefLang.Href, params)
			break
		}
	}
}

// initializeVars creates the variables of the page if needed.
func (page *Page) initializeVars() {
	if page.Content.Vars == nil {
		page.Content.Vars = map[string]any{}
	}
}

// localeVarKey returns the key of the variable name for the locale.
func localeVarKey(name, locale string) string {
	return fmt.Sprintf("%s/%s", name, strings.ToLower(locale))
}

// rawString returns the untranslated message key of the locale of the page, or key if there is none.
func (page *Page) rawString(key string) string {
	rawString := localization.Source(page.Locale, key)

	if len(rawString) == 0 {
		return key
	}

	return rawString
}

// replaceVars executes entry as a template with the content of the page, and returns an empty string on error.
func (page *Page) replaceVars(entry string) string {
	return replaceVars(entry, page.Content)
}

// replaceURLVars executes entry as a template without data, then replaces the parameters, prefixed by a colon, with
// the values of the first map of vars.
func (page *Page) replaceURLVars(entry string, vars ...map[string]string) string {
	url := replaceVars(entry, nil)

	if len(vars) > 0 {
		for key, value := range vars[0] {
			url = strings.Replace(url, ":"+key, value, 1)
		}
	}

	return url
}

// replaceVars executes entry as a template with vars and returns the result,
// or an empty string, after logging the error, if it fails. The data is not logged, since it may be sensitive.
func replaceVars(entry string, vars any) string {
	var output bytes.Buffer

	t, err := template.New("_temp").Parse(entry)
	if err != nil {
		log.WithFields(logrus.Fields{
			"string": entry,
			"error":  err.Error(),
		}).Error("unable to parse the string")
		return ""
	}

	if err := t.Execute(&output, vars); err != nil {
		log.WithFields(logrus.Fields{
			"string": entry,
			"error":  err.Error(),
		}).Error("unable to execute the template")
		return ""
	}

	return output.String()
}
