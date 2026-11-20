package log

import (
	"context"
	"errors"
	"testing"

	commonErrors "github.com/cygy/ginamite/common/errors"
)

func TestDatabaseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"not found", commonErrors.ErrNotFound, false},
		{"other", errors.New("boom"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := useHook(t)
			DatabaseError(tt.err, "users")

			if got := len(hook.AllEntries()) == 1; got != tt.want {
				t.Fatalf("logged = %v, want %v", got, tt.want)
			}
			if tt.want {
				entry := hook.LastEntry()
				if entry.Data["collection"] != "users" || entry.Data["error"] != "boom" {
					t.Errorf("fields = %v", entry.Data)
				}
			}
		})
	}
}

func TestDatabaseOperationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"not found", commonErrors.ErrNotFound, false},
		{"other", errors.New("boom"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := useHook(t)
			ctx := ContextWithAttributes(context.Background(), "request_id", "r1", "route", "/puzzle/:id")
			DatabaseOperationError(ctx, tt.err, "Aggregate", "scores")

			if got := len(hook.AllEntries()) == 1; got != tt.want {
				t.Fatalf("logged = %v, want %v", got, tt.want)
			}
			if !tt.want {
				return
			}

			entry := hook.LastEntry()
			if entry.Message != "database error" || entry.Data["error"] != "boom" {
				t.Errorf("entry = %+v", entry)
			}
			for key, want := range map[string]any{"operation": "Aggregate", "collection": "scores", "request_id": "r1", "route": "/puzzle/:id"} {
				if entry.Data[key] != want {
					t.Errorf("%s = %v, want %v", key, entry.Data[key], want)
				}
			}
		})
	}
}

func TestDatabaseOperationError_NilContext(t *testing.T) {
	hook := useHook(t)

	//nolint:staticcheck // a nil context is tolerated
	DatabaseOperationError(nil, errors.New("boom"), "Find", "users")

	entry := hook.LastEntry()
	if entry == nil || entry.Data["operation"] != "Find" || entry.Data["collection"] != "users" {
		t.Errorf("entry = %+v", entry)
	}
}

func TestPackageOf(t *testing.T) {
	tests := map[string]string{
		"github.com/cygy/ginamite/common/log.DatabaseError": "github.com/cygy/ginamite/common/log",
		"github.com/cygy/app/model.(*Score).Save":           "github.com/cygy/app/model",
		"main.main": "main",
		"main":      "main",
	}

	for function, want := range tests {
		if got := packageOf(function); got != want {
			t.Errorf("packageOf(%q) = %q, want %q", function, got, want)
		}
	}
}
