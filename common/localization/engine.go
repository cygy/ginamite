package localization

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
)

// translationFormats lists the formats of the translation files, by order of preference.
var translationFormats = []string{"toml", "json"}

// Initialize loads the translations of the locales from the files "<path>/<locale>.all.toml" or, if there is none,
// "<path>/<locale>.all.json". It returns an error if a file is missing, invalid or empty.
func (e *Engine) Initialize(path string, locales []string) error {
	for _, locale := range locales {
		lowerLocale := strings.ToLower(locale)

		filePath, format, content, err := readTranslations(path, lowerLocale)
		if err != nil {
			return err
		}

		// One bundle per locale, like with go-i18n v1: a missing key is never
		// resolved with another language.
		bundle := i18n.NewBundle(language.Make(locale))
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

		// go-i18n guesses the language from the file name, and "<locale>.all.<format>"
		// would be read as the language "all". The content is already loaded,
		// so the name passed here is only used to give the language: "<locale>.<format>".
		messageFile, err := bundle.ParseMessageFileBytes(content, lowerLocale+"."+format)
		if err != nil {
			return fmt.Errorf("unable to load the translations %s: %w", filePath, err)
		}

		if len(messageFile.Messages) == 0 {
			return fmt.Errorf("no translation in the file %s", filePath)
		}

		sources := make(map[string]string, len(messageFile.Messages))
		for _, message := range messageFile.Messages {
			sources[message.ID] = message.Other
		}

		e.Translations[locale] = sources
		e.TranslateFuncs[locale] = newTranslateFunc(i18n.NewLocalizer(bundle, locale))
	}

	return nil
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

// readTranslations reads the translation file of the locale in the first format found. It returns the path of the
// file, its format and its content.
func readTranslations(path, lowerLocale string) (filePath, format string, content []byte, err error) {
	for _, format = range translationFormats {
		filePath = fmt.Sprintf("%s/%s.all.%s", path, lowerLocale, format)

		content, err = os.ReadFile(filePath)
		if err == nil {
			return filePath, format, content, nil
		}

		if !errors.Is(err, fs.ErrNotExist) {
			break
		}
	}

	return "", "", nil, fmt.Errorf("unable to read the translations: %w", err)
}
