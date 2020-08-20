package localization

import (
	"fmt"
	"strings"

	"github.com/cygy/ginamite/common/log"
	"github.com/nicksnyder/go-i18n/i18n/bundle"
	"github.com/nicksnyder/go-i18n/i18n/language"
	"github.com/sirupsen/logrus"
)

// Initialize : initalizes the translation engine.
func (e *Engine) Initialize(path string, locales []string) {
	for _, locale := range locales {
		filePath := fmt.Sprintf("%s/%s.all.json", path, strings.ToLower(locale))
		bundle := bundle.New()
		bundle.MustLoadTranslationFile(filePath)

		if translate, err := bundle.Tfunc(locale); err != nil {
			log.WithFields(logrus.Fields{
				"path":  filePath,
				"error": err.Error(),
			}).Panic("unable to load translation file")
		} else {
			e.TranslateFuncs[locale] = translate
		}

		for locale, mapTranslate := range bundle.Translations() {
			e.Translations[locale] = mapTranslate
		}
	}
}

// Translate : returns the right translate function.
func (e *Engine) Translate(locale string) bundle.TranslateFunc {
	translateFunc, ok := e.TranslateFuncs[locale]
	if !ok {
		return nil
	}

	return translateFunc
}

// Source : returns the raw source string.
func (e *Engine) Source(locale, key string) string {
	translationMap, ok := e.Translations[locale]
	if !ok {
		return ""
	}

	translation, ok := translationMap[key]
	if !ok {
		return ""
	}

	return translation.Template(language.Other).String()
}
