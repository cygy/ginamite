package authtoken

import (
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// hooksContextKey is the key of the hooks in the gin context.
const hooksContextKey = "_authtoken_hooks"

// InjectHooks returns a middleware saving the hooks that build and extend the auth tokens in the gin context,
// see HooksOf.
func InjectHooks(hooks authentication.Hooks) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetHooks(c, hooks)
		c.Next()
	}
}

// SetHooks saves the hooks in the gin context, like InjectHooks does. It is meant for the tests.
func SetHooks(c *gin.Context, hooks authentication.Hooks) {
	c.Set(hooksContextKey, hooks)
}

// HooksOf returns the hooks saved in the gin context by InjectHooks.
// It panics if there are none: the handlers can not sign the auth tokens without them.
func HooksOf(c *gin.Context) authentication.Hooks {
	value, ok := c.Get(hooksContextKey)
	if !ok {
		panic("authtoken: no hooks in the context, add the middleware InjectHooks")
	}

	return value.(authentication.Hooks)
}

// revokedContextKey is the key of the function reporting the revoked tokens in the gin context.
const revokedContextKey = "_authtoken_revoked"

// IsRevokedFunc reports whether an auth token is revoked.
type IsRevokedFunc func(tokenID string) bool

// InjectRevoked returns a middleware saving in the gin context the function reporting whether an auth token is
// revoked, which is used by Parse. Without it no token is revoked.
func InjectRevoked(isRevoked IsRevokedFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetRevoked(c, isRevoked)
		c.Next()
	}
}

// SetRevoked saves the function in the gin context, like InjectRevoked does. It is meant for the tests.
func SetRevoked(c *gin.Context, isRevoked IsRevokedFunc) {
	c.Set(revokedContextKey, isRevoked)
}

// isRevoked reports whether the auth token is revoked according to the function of the gin context, if any.
func isRevoked(c *gin.Context, tokenID string) bool {
	if value, ok := c.Get(revokedContextKey); ok {
		return value.(IsRevokedFunc)(tokenID)
	}

	return false
}
