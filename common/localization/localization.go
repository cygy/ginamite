// Package localization translates messages by locale, from go-i18n translation files.
package localization

// H is a shorthand for a map[string]any, used as template data.
type H map[string]any

// S is a shorthand for a map[string]string.
type S map[string]string

// TranslateFunc translates a message ID.
// The optional arguments follow the go-i18n v1 convention: a number as first
// argument is the plural count, and the next argument (or the first one if
// there is no count) is the template data (map or struct).
type TranslateFunc func(messageID string, args ...any) string

// Engine translates messages for a set of locales.
type Engine struct {
	// TranslateFuncs holds one translate function per locale.
	TranslateFuncs map[string]TranslateFunc
	// Translations holds the raw source strings, by locale and by message ID.
	Translations map[string]map[string]string
}

// NewEngine returns an empty Engine.
func NewEngine() (e *Engine) {
	e = &Engine{
		TranslateFuncs: map[string]TranslateFunc{},
		Translations:   map[string]map[string]string{},
	}

	return
}

// Load returns an Engine with the translations of the locales loaded from the directory path.
// See Engine.Initialize.
func Load(path string, locales []string) *Engine {
	e := NewEngine()
	e.Initialize(path, locales)

	return e
}
