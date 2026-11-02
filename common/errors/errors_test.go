package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNotFound(t *testing.T) {
	err := NotFound()
	if !errors.Is(err, ErrNotFound) || err.Error() != NotFoundErrorMessage {
		t.Errorf("NotFound() = %v, want ErrNotFound", err)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"not found", NotFound(), true},
		{"wrapped", fmt.Errorf("user 42: %w", ErrNotFound), true},
		{"same text", errors.New(NotFoundErrorMessage), false},
		{"other error", errors.New("boom"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
