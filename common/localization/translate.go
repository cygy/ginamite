package localization

import (
	"reflect"
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// newTranslateFunc returns a function with the same behavior as the go-i18n v1 one.
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
			config.PluralCount = pluralCount(count)
		}

		translation, err := localizer.Localize(config)
		if err != nil && translation == "" {
			// v1 returned the ID when the translation was missing.
			return messageID
		}

		return translation
	}
}

// pluralCount returns the count in a form go-i18n accepts: floats and unsigned integers must be given as strings.
func pluralCount(count any) any {
	switch v := count.(type) {
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	}

	return count
}

// isNumber reports whether n is a number, or a string holding one.
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

// toMap converts a map (with string keys) or a struct to a new map[string]any.
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
