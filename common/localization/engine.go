package localization

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cygy/ginamite/common/log"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

// Initialize loads the translations of the locales from the files "<path>/<locale>.all.json".
// It panics, after logging the error, if a file is missing, invalid or empty.
func (e *Engine) Initialize(path string, locales []string) {
	for _, locale := range locales {
		lowerLocale := strings.ToLower(locale)
		filePath := fmt.Sprintf("%s/%s.all.json", path, lowerLocale)

		content, err := os.ReadFile(filePath)
		if err != nil {
			log.WithFields(logrus.Fields{
				"path":  filePath,
				"error": err.Error(),
			}).Panic("unable to read translation file")
		}

		// One bundle per locale, like with go-i18n v1: a missing key is never
		// resolved with another language.
		bundle := i18n.NewBundle(language.Make(locale))
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

		// go-i18n guesses the language from the file name, and "<locale>.all.json"
		// would be read as the language "all". The content is already loaded,
		// so the name passed here is only used to give the language: "<locale>.json".
		messageFile, err := bundle.ParseMessageFileBytes(content, lowerLocale+".json")
		if err != nil {
			log.WithFields(logrus.Fields{
				"path":  filePath,
				"error": err.Error(),
			}).Panic("unable to load translation file")
		}

		if len(messageFile.Messages) == 0 {
			log.WithFields(logrus.Fields{
				"path":   filePath,
				"locale": locale,
			}).Panic("no translation found in the translation file")
		}

		sources := make(map[string]string, len(messageFile.Messages))
		for _, message := range messageFile.Messages {
			sources[message.ID] = message.Other
		}

		e.Translations[locale] = sources
		e.TranslateFuncs[locale] = newTranslateFunc(i18n.NewLocalizer(bundle, locale))
	}
}

// Translate returns the translate function of the locale, or nil if the locale is not loaded.
func (e *Engine) Translate(locale string) TranslateFunc {
	translateFunc, ok := e.TranslateFuncs[locale]
	if !ok {
		return nil
	}

	return translateFunc
}

// Source returns the raw, untranslated source string of the message key for the locale,
// or an empty string if it does not exist.
func (e *Engine) Source(locale, key string) string {
	return e.Translations[locale][key]
}
