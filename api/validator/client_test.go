package validator

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIClient(t *testing.T) {
	var accept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := apiClient(server.URL, 1500, 2, false)

	if client.RetryCount != 2 {
		t.Errorf("RetryCount = %d, want 2", client.RetryCount)
	}
	if want := 1500 * time.Millisecond; client.GetClient().Timeout != want {
		t.Errorf("Timeout = %v, want %v", client.GetClient().Timeout, want)
	}
	if _, err := client.R().Get("/"); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if accept != "application/json" {
		t.Errorf("Accept header = %q, want application/json", accept)
	}
}

func TestExtractClaimsFromJWTToken(t *testing.T) {
	validator := NewGoogleTokenValidator("id", "v3", 1000, 0, false)
	payload := func(s string) string { return "a." + base64.StdEncoding.EncodeToString([]byte(s)) + ".c" }

	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"valid", payload(`{"sub":"user"}`), true},
		{"wrong number of segments", "a.b", false},
		{"payload not base64", "a.%%%.c", false},
		{"payload not json", payload("nope"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var claims struct {
				Sub string `json:"sub"`
			}

			if got := extractClaimsFromJWTToken(tt.token, &claims, validator); got != tt.want {
				t.Fatalf("extractClaimsFromJWTToken() = %v, want %v", got, tt.want)
			}
			if tt.want && claims.Sub != "user" {
				t.Errorf("Sub = %q, want user", claims.Sub)
			}
		})
	}
}

func TestLogResponseToken(t *testing.T) {
	tests := []struct {
		name   string
		status int
		err    error
		want   bool
	}{
		{"ok", http.StatusOK, nil, true},
		{"bad status", http.StatusBadRequest, nil, false},
		{"request error", http.StatusOK, errors.New("boom"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			response, err := apiClient(server.URL, 1000, 0, false).R().Get("/")
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			validator := NewFacebookTokenValidator("id", "v3", 1000, 0, false)
			if got := logResponseToken("token", validator, response, tt.err); got != tt.want {
				t.Errorf("logResponseToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogParseToken(t *testing.T) {
	validator := NewGoogleTokenValidator("id", "v3", 1000, 0, false)

	var value map[string]any
	if !logParseToken("token", validator, []byte(`{"a":1}`), &value) || value["a"] == nil {
		t.Errorf("logParseToken() must parse a valid payload, got %v", value)
	}
	if logParseToken("token", validator, []byte("nope"), &value) {
		t.Error("logParseToken() = true, want false for an invalid payload")
	}
}
