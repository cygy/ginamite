package html

import (
	"fmt"
	"strings"
)

// Async is a passthrough asynchronous call: the web server forwards the request of the browser to the API.
type Async struct {
	Paths  map[string]string // The path of the call, by locale.
	Method string            // The HTTP method of the call.
	Target string            // The endpoint of the API, with the parameters of the path prefixed by a colon.
}

// Prepare keeps the paths of the supported locales, prefixes them with webHost and replaces the locale keyword.
func (async *Async) Prepare(webHost string, applicationLocales []string) {
	paths := map[string]string{}
	localeKey := fmt.Sprintf("{{.%s}}", LocaleKeyword)
	for _, applicationLocale := range applicationLocales {
		if path, ok := async.Paths[applicationLocale]; ok {
			paths[applicationLocale] = strings.Replace(webHost+path, localeKey, applicationLocale, 1)
		}
	}
	async.Paths = paths
}
