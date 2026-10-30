package errors

import (
	"errors"
	"testing"
)

func TestNotFound(t *testing.T) {
	err := NotFound()
	if err == nil || err.Error() != NotFoundErrorMessage {
		t.Errorf("NotFound() = %v, want %q", err, NotFoundErrorMessage)
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
		{"different case", errors.New("Not Found"), true},
		{"other error", errors.New("boom"), false},
		{"wrapped", errors.New("x: not found"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
