package request

import "testing"

func TestIsValidID(t *testing.T) {
	for _, id := range []string{"", "xyz", "507f1f77bcf86cd799439011"} {
		if !IsValidID(id) {
			t.Errorf("IsValidID(%q) = false, want true by default", id)
		}
	}
}

// setIDValidator replaces IsValidID for the duration of the test.
func setIDValidator(t *testing.T, validator func(string) bool) {
	t.Helper()

	old := IsValidID
	IsValidID = validator
	t.Cleanup(func() { IsValidID = old })
}
