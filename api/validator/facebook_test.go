package validator

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/authentication"
)

func TestFacebookTokenValidator(t *testing.T) {
	tests := []struct {
		name       string
		debugBody  string
		debugCode  int
		meBody     string
		meCode     int
		wantResult bool
	}{
		{"valid token", `{"data":{"app_id":"app-id","expires_at":2000,"is_valid":true}}`, 200, `{"id":"fb-user","email":"john@example.com","first_name":"John","last_name":"Doe"}`, 200, true},
		{"wrong application", `{"data":{"app_id":"other","expires_at":2000,"is_valid":true}}`, 200, `{}`, 200, false},
		{"invalid token", `{"data":{"app_id":"app-id","is_valid":false}}`, 200, `{}`, 200, false},
		{"debug rejected", `{}`, 400, `{}`, 200, false},
		{"debug not json", `nope`, 200, `{}`, 200, false},
		{"profile rejected", `{"data":{"app_id":"app-id","expires_at":2000,"is_valid":true}}`, 200, `{}`, 500, false},
		{"profile not json", `{"data":{"app_id":"app-id","expires_at":2000,"is_valid":true}}`, 200, `nope`, 200, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/debug_token", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.debugCode)
				_, _ = w.Write([]byte(tt.debugBody))
			})
			mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.meCode)
				_, _ = w.Write([]byte(tt.meBody))
			})
			server := httptest.NewServer(mux)
			defer server.Close()

			validator := &facebookTokenValidator{client: apiClient(server.URL+"/", 1000, 0, false), applicationID: "app-id"}
			infos, ok := validator.VerifyTokenAndExtractInfos("token")

			if ok != tt.wantResult {
				t.Fatalf("ok = %v, want %v", ok, tt.wantResult)
			}
			if ok && (infos.UserID != "fb-user" || infos.Email != "john@example.com" || !infos.ExpiresAt.Equal(time.Unix(2000, 0))) {
				t.Errorf("infos = %+v, want the profile of the user", infos)
			}
		})
	}
}

func TestFacebookTokenValidator_SourceName(t *testing.T) {
	if got := NewFacebookTokenValidator("id", "v3", 1000, 0, false).SourceName(); got != authentication.MethodFacebook {
		t.Errorf("source = %q, want %q", got, authentication.MethodFacebook)
	}
}
