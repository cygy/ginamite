package response

import (
	"encoding/json"
	"testing"
)

func TestError_Error(t *testing.T) {
	e := &Error{Message: "m", Reason: "r", Recovery: "c"}
	if got := e.Error(); got != "m r c" {
		t.Errorf("Error() = %q", got)
	}
}

func TestError_IsNil(t *testing.T) {
	tests := []struct {
		name string
		err  Error
		want bool
	}{
		{"empty", Error{}, true},
		{"only a message", Error{Message: "m"}, true},
		{"with a code", Error{Code: 1}, false},
		{"with a domain", Error{Domain: "d"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsNil(); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_JSON(t *testing.T) {
	data, err := json.Marshal(Error{Code: 2, Domain: "d"})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"code":2,"domain":"d"}` {
		t.Errorf("json = %s", data)
	}
}
