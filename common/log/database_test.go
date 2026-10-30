package log

import (
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
		{"not found", commonErrors.NotFound(), false},
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
