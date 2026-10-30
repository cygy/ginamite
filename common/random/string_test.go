package random

import (
	"strings"
	"testing"
)

func TestString_Length(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"negative", -3, 0},
		{"zero", 0, 0},
		{"one", 1, 1},
		{"sixteen", 16, 16},
		{"long", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(String(tt.n)); got != tt.want {
				t.Errorf("len(String(%d)) = %d, want %d", tt.n, got, tt.want)
			}
		})
	}
}

func TestString_Alphabet(t *testing.T) {
	for _, r := range String(2000) {
		if !strings.ContainsRune(string(randomChars), r) {
			t.Fatalf("unexpected character %q", r)
		}
	}
}

func TestString_Varies(t *testing.T) {
	first, second := String(32), String(32)
	if first == second {
		t.Error("two random strings of 32 characters must differ")
	}
}
