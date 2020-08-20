package route

import (
	"strings"

	"github.com/cygy/ginamite/web/html"
)

// Returns the relative path of an url given a locale and the key of that url (from the route configuration file).
func relativeRoutePath(key, locale string) (string, bool, bool) {
	url := html.Main.URL(key, locale)
	partial := html.Main.IsPartial(key, locale)
	path, ok := relativePath(url, locale)

	return path, partial, ok
}

// Returns the relative path of a async url given a locale and the key of that url (from the route configuration file).
func relativeAsyncPath(key, locale string) (string, bool) {
	url := html.Main.AsyncURL(key, locale)
	return relativePath(url, locale)
}

// Returns the relative path of a async url given a locale and the key of that url (from the route configuration file).
func relativePath(url, locale string) (string, bool) {
	if parts := strings.Split(url, "/"+locale); len(parts) == 2 {
		return parts[1], true
	}

	return "", false
}
