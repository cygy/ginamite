package localization

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/cygy/ginamite/common/log"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

// Initialize : initalizes the translation engine.
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

// Translate : returns the right translate function.
func (e *Engine) Translate(locale string) TranslateFunc {
	translateFunc, ok := e.TranslateFuncs[locale]
	if !ok {
		return nil
	}

	return translateFunc
}

// Source : returns the raw source string.
func (e *Engine) Source(locale, key string) string {
	return e.Translations[locale][key]
}

// newTranslateFunc : returns a function with the same behavior as the go-i18n v1 one.
func newTranslateFunc(localizer *i18n.Localizer) TranslateFunc {
	return func(messageID string, args ...any) string {
		var count, data any

		if len(args) > 0 {
			if isNumber(args[0]) {
				count = args[0]
				if len(args) > 1 {
					data = args[1]
				}
			} else {
				data = args[0]
			}
		}

		// v1 exposed the count to the templates as {{.Count}}, and also accepted
		// a "Count" key in the data to select the plural form.
		templateData := toMap(data)
		if count != nil {
			templateData["Count"] = count
		} else if c, ok := templateData["Count"]; ok {
			count = c
		}

		config := &i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: templateData,
		}
		if count != nil {
			config.PluralCount = count
		}

		translation, err := localizer.Localize(config)
		if err != nil && translation == "" {
			// v1 returned the ID when the translation was missing.
			return messageID
		}

		return translation
	}
}

func isNumber(n any) bool {
	switch v := n.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case string:
		_, err := strconv.ParseFloat(v, 64)
		return err == nil
	}

	return false
}

// toMap : converts a map (with string keys) or a struct to a new map[string]interface{}.
func toMap(data any) map[string]any {
	result := map[string]any{}
	if data == nil {
		return result
	}

	value := reflect.ValueOf(data)
	for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return result
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return result
		}
		for _, key := range value.MapKeys() {
			result[key.String()] = value.MapIndex(key).Interface()
		}
	case reflect.Struct:
		valueType := value.Type()
		for i := 0; i < value.NumField(); i++ {
			if field := valueType.Field(i); field.PkgPath == "" {
				result[field.Name] = value.Field(i).Interface()
			}
		}
	}

	return result
}
