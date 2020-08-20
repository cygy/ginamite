package localization

import (
	"github.com/nicksnyder/go-i18n/i18n/bundle"
	"github.com/nicksnyder/go-i18n/i18n/translation"
)

// H : shorthand to use map[string]interface{}
type H map[string]interface{}

// S : shorthand to use map[string]string
type S map[string]string

// Engine :
type Engine struct {
	TranslateFuncs map[string]bundle.TranslateFunc
	Translations   map[string]map[string]translation.Translation
}

// NewEngine : returns a new Engine struct.
func NewEngine() (e *Engine) {
	e = &Engine{
		TranslateFuncs: map[string]bundle.TranslateFunc{},
		Translations:   map[string]map[string]translation.Translation{},
	}

	return
}
