package route

import (
	"strings"

	"github.com/cygy/ginamite/web/html"
)

// relativeRoutePath returns the path of the route key for the locale relative to the group of the locale, whether the
// page is partial, and whether the route exists for the locale.
func relativeRoutePath(pages *html.Configuration, key, locale string) (string, bool, bool) {
	path, ok := relativePath(pages.URL(key, locale), locale)

	return path, pages.IsPartial(key, locale), ok
}

// relativeAsyncPath returns the path of the async call key for the locale relative to the group of the locale, and
// whether the call exists for the locale.
func relativeAsyncPath(pages *html.Configuration, key, locale string) (string, bool) {
	return relativePath(pages.AsyncURL(key, locale), locale)
}

// relativePath returns the part of the URL after the locale, and whether the URL has exactly one occurrence of it.
func relativePath(url, locale string) (string, bool) {
	if parts := strings.Split(url, "/"+locale); len(parts) == 2 {
		return parts[1], true
	}

	return "", false
}
