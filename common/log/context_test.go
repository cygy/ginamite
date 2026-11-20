package log

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
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

func TestAttributesFromContext_GinContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ContextWithAttributes(req.Context(), "request_id", "r1"))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	if got, want := AttributesFromContext(c), []any{"request_id", "r1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("attributes = %v, want %v", got, want)
	}

	// A gin context without request has no attribute.
	c.Request = nil
	if got := AttributesFromContext(c); got != nil {
		t.Errorf("attributes without request = %v", got)
	}
}
