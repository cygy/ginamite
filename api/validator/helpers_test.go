package validator

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func jwtWithClaims(t *testing.T, claims map[string]any) string {
	t.Helper()

	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("unable to encode the claims: %v", err)
	}

	return "header." + base64.StdEncoding.EncodeToString(payload) + ".signature"
}
