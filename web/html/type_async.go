package html

import (
	"fmt"
	"strings"
)

// Async : struct of a passthrough async call.
type Async struct {
	Paths  map[string]string
	Method string
	Target string
}

// Prepare : sets up the localized urls.
func (async *Async) Prepare(webHost string, applicationLocales []string) {
	// Remove the unsupported locales from the paths map.
	paths := map[string]string{}
	localeKey := fmt.Sprintf("{{.%s}}", LocaleKeyword)
	for _, applicationLocale := range applicationLocales {
		if path, ok := async.Paths[applicationLocale]; ok {
			paths[applicationLocale] = strings.Replace(webHost+path, localeKey, applicationLocale, 1)
		}
	}
	async.Paths = paths
}
