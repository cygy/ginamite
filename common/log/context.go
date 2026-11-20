package log

import (
	"context"

	"github.com/gin-gonic/gin"
)

// attributesKey is the key of the log attributes in a context.
type attributesKey struct{}

// ContextWithAttributes returns a copy of ctx carrying the attributes, alternating keys and values as in
// slog.Logger.Log, after the ones it already carries. They are added to the records which log with the context, such
// as DatabaseOperationError: InjectRequestLogger saves the ID, the method and the route of each request, and a handler
// can add what identifies its work, such as the ID of the puzzle.
func ContextWithAttributes(ctx context.Context, args ...any) context.Context {
	if len(args) == 0 {
		return ctx
	}

	existing := AttributesFromContext(ctx)
	attrs := make([]any, 0, len(existing)+len(args))
	attrs = append(attrs, existing...)
	attrs = append(attrs, args...)

	return context.WithValue(ctx, attributesKey{}, attrs)
}

// AttributesFromContext returns the attributes saved in ctx by ContextWithAttributes, or nil if there is none. The
// context of the request of a *gin.Context is read too: the handlers often pass the gin context itself to the storage.
//
//nolint:contextcheck // the context of the request of a gin context is read on purpose
func AttributesFromContext(ctx context.Context) []any {
	if ctx == nil {
		return nil
	}

	if c, ok := ctx.(*gin.Context); ok && c.Request != nil {
		ctx = c.Request.Context()
	}

	attrs, _ := ctx.Value(attributesKey{}).([]any)

	return attrs
}
