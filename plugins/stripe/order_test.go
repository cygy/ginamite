package stripe

import "testing"

func TestOrder_Total(t *testing.T) {
	order := Order{Items: []Item{{Amount: 499, Quantity: 2}, {Amount: 100, Quantity: 1}}}
	if got := order.total(); got != 1098 {
		t.Errorf("total() = %d, want 1098", got)
	}
	if got := (Order{}).total(); got != 0 {
		t.Errorf("total() = %d, want 0", got)
	}
}

func TestOrder_Validate(t *testing.T) {
	tests := map[string]struct {
		order Order
		ok    bool
	}{
		"valid":       {Order{Items: []Item{{Name: "Pack", Amount: 499, Quantity: 1}}}, true},
		"no item":     {Order{}, false},
		"free":        {Order{Items: []Item{{Name: "Pack", Amount: 0, Quantity: 1}}}, false},
		"no quantity": {Order{Items: []Item{{Name: "Pack", Amount: 499}}}, false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tt.order.validate(); (err == nil) != tt.ok {
				t.Errorf("validate() = %v, want ok = %v", err, tt.ok)
			}
		})
	}
}
