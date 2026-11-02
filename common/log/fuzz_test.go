package log

import (
	"encoding/json"
	"strings"
	"testing"
)

// FuzzObfuscateBody checks that the value of a secret key never reaches the logs, and that no body nor query makes the
// obfuscation panic.
func FuzzObfuscateBody(f *testing.F) {
	f.Add("password", "secret", `{"a":1}`)
	f.Add("auth_token", `a"b\c`, `[]`)
	f.Add("Client_Secret", "é€", `{"email":"john@example.com"}`)

	f.Fuzz(func(t *testing.T, key, value, other string) {
		obfuscateQuery(other)
		obfuscateBody([]byte(other))

		if !sensitiveKeyRegExp.MatchString(key) {
			return
		}

		// The body holds the secret only: once masked, its JSON value is gone.
		body := []byte("{" + string(mustMarshal(t, key)) + ":" + string(mustMarshal(t, value)) + "}")
		if masked := obfuscateBody(body); !strings.HasSuffix(string(masked), `:"***"}`) {
			t.Errorf("the secret %q of %q is logged: %s", value, key, masked)
		}
	})
}

// mustMarshal returns the JSON string of s, as it appears in a body.
func mustMarshal(t *testing.T, s string) []byte {
	t.Helper()

	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	return b
}
