package log

import "context"

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

// AttributesFromContext returns the attributes saved in ctx by ContextWithAttributes, or nil if there is none.
func AttributesFromContext(ctx context.Context) []any {
	if ctx == nil {
		return nil
	}

	attrs, _ := ctx.Value(attributesKey{}).([]any)

	return attrs
}
