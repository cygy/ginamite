package functions

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

func TestDefaultDocumentsWrappers(t *testing.T) {
	tests := []struct {
		name string
		call func(ctx context.Context, fns *Functions, f DefaultDocumentsHook)
	}{
		{"CreateDefaultDocuments", func(ctx context.Context, fns *Functions, f DefaultDocumentsHook) {
			fns.CreateDefaultDocuments(ctx, "test", "1.0", f)
		}},
		{"RemoveDefaultDocuments", func(ctx context.Context, fns *Functions, f DefaultDocumentsHook) {
			fns.RemoveDefaultDocuments(ctx, "test", "1.0", f)
		}},
		{"UpdateDefaultDocuments", func(ctx context.Context, fns *Functions, f DefaultDocumentsHook) {
			fns.UpdateDefaultDocuments(ctx, "test", "1.0", f)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fns, db := withMainSession(t)

			type key struct{}
			ctx := context.WithValue(context.Background(), key{}, "worker")

			var calls int
			tt.call(ctx, fns, func(gotCtx context.Context, environment, version string, got store.Database) {
				calls++
				if gotCtx.Value(key{}) != "worker" || environment != "test" || version != "1.0" || got != store.Database(db) {
					t.Errorf("arguments = %v, %q, %q, %v", gotCtx, environment, version, got)
				}
			})
			if calls != 1 {
				t.Errorf("%d calls, want 1", calls)
			}
		})
	}
}

func TestDefaultDocumentsOfThePlugin(t *testing.T) {
	// The plugin has no default documents: nothing is done, and nothing needs the database.
	CreateAllDefaultDocuments("test", "1.0")
	RemoveUnusedDefaultDocuments("test", "1.0")
	UpdateAllDefaultDocuments("test", "1.0")
}
