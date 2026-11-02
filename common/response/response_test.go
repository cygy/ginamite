package response

import (
	"encoding/json"
	"testing"
)

func TestResponse_IsErrorNil(t *testing.T) {
	tests := []struct {
		name string
		r    Response
		want bool
	}{
		{"no error", Response{Status: "ok"}, true},
		{"empty error", Response{Error: &Error{}}, true},
		{"error", Response{Error: &Error{Code: 1, Domain: "d"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsErrorNil(); got != tt.want {
				t.Errorf("IsErrorNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	r := NewError(Error{Code: 3, Domain: "d"})
	if r.Error == nil || r.Error.Code != 3 || r.Status != "" || r.ID != "" {
		t.Errorf("NewError() = %+v", r)
	}
}

func TestNewStatus(t *testing.T) {
	if r := NewStatus("done"); r.Status != "done" || r.Error != nil || r.ID != "" {
		t.Errorf("NewStatus() = %+v", r)
	}
}

func TestNewStatusAndID(t *testing.T) {
	r := NewStatusAndID("done", "42")

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"status":"done","id":"42"}` {
		t.Errorf("json = %s", data)
	}
}
