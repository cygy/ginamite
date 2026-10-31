package functions

import (
	"testing"

	"github.com/cygy/ginamite/common/mongo/store"
)

func TestDefaultDocumentsWrappers(t *testing.T) {
	tests := []struct {
		name string
		call func(fns *Functions, f func(string, string, store.Database))
	}{
		{"CreateDefaultDocuments", func(fns *Functions, f func(string, string, store.Database)) {
			fns.CreateDefaultDocuments("test", "1.0", f)
		}},
		{"RemoveDefaultDocuments", func(fns *Functions, f func(string, string, store.Database)) {
			fns.RemoveDefaultDocuments("test", "1.0", f)
		}},
		{"UpdateDefaultDocuments", func(fns *Functions, f func(string, string, store.Database)) {
			fns.UpdateDefaultDocuments("test", "1.0", f)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fns, db := withMainSession(t)

			var calls int
			tt.call(fns, func(environment, version string, got store.Database) {
				calls++
				if environment != "test" || version != "1.0" || got != store.Database(db) {
					t.Errorf("arguments = %q, %q, %v", environment, version, got)
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
