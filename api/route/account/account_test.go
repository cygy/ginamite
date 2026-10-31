package account

import (
	"net/http"
	"strings"
	"testing"
)

func TestCheckUsernameParameter(t *testing.T) {
	tests := []struct {
		name     string
		username string
		forbid   []string
		want     bool
		status   int
	}{
		{"empty", "", nil, false, http.StatusBadRequest},
		{"too short", "ab", nil, false, http.StatusBadRequest},
		{"invalid characters", "john/doe", nil, false, http.StatusBadRequest},
		{"forbidden string", "superAdmin42", []string{"admin"}, false, http.StatusPreconditionFailed},
		{"consonants only", "bcdfghjk", nil, false, http.StatusPreconditionFailed},
		{"random case", "tytHRBCWGz", nil, false, http.StatusPreconditionFailed},
		{"valid", "Cyril", nil, true, http.StatusOK},
		{"valid long", "AvianAnalyst", nil, true, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			testConfig.Account.InvalidStringsForUsername = tt.forbid

			c, recorder := newContext(nil)
			if got := CheckUsernameParameter(c, tt.username, "username"); got != tt.want {
				t.Fatalf("CheckUsernameParameter() = %v, want %v", got, tt.want)
			}
			c.Writer.WriteHeaderNow()
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d", recorder.Code, tt.status)
			}
		})
	}
}

func TestDoesUsernameContainForbiddenStrings(t *testing.T) {
	setup(t)
	testConfig.Account.InvalidStringsForUsername = []string{"Admin", "root"}

	tests := []struct {
		username string
		want     bool
	}{
		{"john", false},
		{"theADMIN", true},
		{"ROOTuser", true},
	}
	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			c, _ := newContext(nil)
			if got := doesUsernameContainForbiddenStrings(c, tt.username); got != tt.want {
				t.Errorf("doesUsernameContainForbiddenStrings(%q) = %v, want %v", tt.username, got, tt.want)
			}
		})
	}
}

func TestIsUsernamePatternAccepted(t *testing.T) {
	accepted := []string{"Cyril", "keitaro", "AvianAnalyst", "MartinEdush", "theoLord", "ScottLat", "cloudcolors", "FredLand"}
	rejected := []string{"tytHRBCWGz", "CbIVUrts", "bTrAPIQFaIyQv", "zhZmjyKf", "XnTfybBSz", "wCOdjEca", strings.Repeat("b", 8)}

	for _, username := range accepted {
		t.Run("accepted/"+username, func(t *testing.T) {
			if !isUsernamePatternAccepted(username) {
				t.Errorf("isUsernamePatternAccepted(%q) = false, want true", username)
			}
		})
	}
	for _, username := range rejected {
		t.Run("rejected/"+username, func(t *testing.T) {
			if isUsernamePatternAccepted(username) {
				t.Errorf("isUsernamePatternAccepted(%q) = true, want false", username)
			}
		})
	}
}
