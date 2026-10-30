package image

import "testing"

func TestSizes_IsDefined(t *testing.T) {
	tests := []struct {
		name  string
		sizes Sizes
		want  bool
	}{
		{"none", Sizes{}, false},
		{"only 1x", Sizes{Size1x: Property{Name: "a"}}, false},
		{"only 2x", Sizes{Size2x: Property{Name: "a"}}, false},
		{"both", Sizes{Size1x: Property{Name: "a"}, Size2x: Property{Name: "b"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sizes.IsDefined(); got != tt.want {
				t.Errorf("IsDefined() = %v, want %v", got, tt.want)
			}
		})
	}
}
