package localization

// H : shorthand to use map[string]any
type H map[string]any

// S : shorthand to use map[string]string
type S map[string]string

// TranslateFunc : translates a message ID.
// The optional arguments follow the go-i18n v1 convention: a number as first
// argument is the plural count, and the next argument (or the first one if
// there is no count) is the template data (map or struct).
type TranslateFunc func(messageID string, args ...any) string

// Engine :
type Engine struct {
	// TranslateFuncs : one translate function per locale.
	TranslateFuncs map[string]TranslateFunc
	// Translations : raw source strings, by locale and by message ID.
	Translations map[string]map[string]string
}

// NewEngine : returns a new Engine struct.
func NewEngine() (e *Engine) {
	e = &Engine{
		TranslateFuncs: map[string]TranslateFunc{},
		Translations:   map[string]map[string]string{},
	}

	return
}
