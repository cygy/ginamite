package validator

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/authentication"
)

func TestGoogleTokenValidator(t *testing.T) {
	validClaims := map[string]any{
		"aud": "client-id", "sub": "google-user", "email": "john@example.com", "email_verified": true,
		"given_name": "John", "family_name": "Doe", "iat": 1000, "exp": 2000,
	}
	with := func(key string, value any) map[string]any {
		claims := map[string]any{}
		for k, v := range validClaims {
			claims[k] = v
		}
		claims[key] = value

		return claims
	}

	tests := []struct {
		name       string
		token      string
		status     int
		wantResult bool
	}{
		{"valid token", jwtWithClaims(t, validClaims), http.StatusOK, true},
		{"wrong audience", jwtWithClaims(t, with("aud", "other")), http.StatusOK, false},
		{"email not verified", jwtWithClaims(t, with("email_verified", false)), http.StatusOK, false},
		{"missing email", jwtWithClaims(t, with("email", "")), http.StatusOK, false},
		{"missing subject", jwtWithClaims(t, with("sub", "")), http.StatusOK, false},
		{"not a JWT", "not-a-jwt", http.StatusOK, false},
		{"payload not base64", "a.%%%.c", http.StatusOK, false},
		{"payload not json", "a." + base64.StdEncoding.EncodeToString([]byte("nope")) + ".c", http.StatusOK, false},
		{"rejected by the provider", jwtWithClaims(t, validClaims), http.StatusBadRequest, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			validator := &googleTokenValidator{client: apiClient(server.URL+"/", 1000, 0, false), clientID: "client-id"}
			infos, ok := validator.VerifyTokenAndExtractInfos(tt.token)

			if ok != tt.wantResult {
				t.Fatalf("ok = %v, want %v", ok, tt.wantResult)
			}
			if ok {
				if infos.UserID != "google-user" || infos.Email != "john@example.com" || infos.FirstName != "John" || infos.LastName != "Doe" {
					t.Errorf("infos = %+v, want the claims of the token", infos)
				}
				if !infos.CreatedAt.Equal(time.Unix(1000, 0)) || !infos.ExpiresAt.Equal(time.Unix(2000, 0)) {
					t.Errorf("dates = %v, %v, want the claims of the token", infos.CreatedAt, infos.ExpiresAt)
				}
			}
		})
	}
}

func TestGoogleTokenValidator_SourceName(t *testing.T) {
	if got := NewGoogleTokenValidator("id", "v3", 1000, 0, false).SourceName(); got != authentication.MethodGoogle {
		t.Errorf("source = %q, want %q", got, authentication.MethodGoogle)
	}
}
