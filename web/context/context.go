// Package context saves to, and reads from, the Gin context of a request the properties used by the web server: the
// locale, the key of the target page and the authentication token.
package context

const (
	localeKey  = "_locale"
	tokenKey   = "_token"
	pageKeyKey = "_pagekey"
)
