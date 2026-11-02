package rights

import (
	"testing"
)

func TestHasUserAbility(t *testing.T) {
	tests := []struct {
		name      string
		required  string
		abilities []string
		want      bool
	}{
		{"present", "a", []string{"a", "b"}, true},
		{"absent", "c", []string{"a", "b"}, false},
		{"no abilities", "a", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasUserAbility(tt.required, tt.abilities); got != tt.want {
				t.Errorf("hasUserAbility() = %v, want %v", got, tt.want)
			}
		})
	}
}
