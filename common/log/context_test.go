package log

import (
	"context"
	"reflect"
	"testing"
)

func TestContextWithAttributes(t *testing.T) {
	if got := AttributesFromContext(context.Background()); got != nil {
		t.Errorf("attributes of an empty context = %v", got)
	}
	//nolint:staticcheck // a nil context is tolerated by the logging helpers
	if got := AttributesFromContext(nil); got != nil {
		t.Errorf("attributes of a nil context = %v", got)
	}

	ctx := context.Background()
	if got := ContextWithAttributes(ctx); got != ctx {
		t.Error("no attribute must return the same context")
	}

	first := ContextWithAttributes(ctx, "request_id", "r1")
	second := ContextWithAttributes(first, "puzzle", "p1")

	if got, want := AttributesFromContext(second), []any{"request_id", "r1", "puzzle", "p1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("attributes = %v, want %v", got, want)
	}
	if got, want := AttributesFromContext(first), []any{"request_id", "r1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("the parent context must keep its attributes: %v, want %v", got, want)
	}
}
