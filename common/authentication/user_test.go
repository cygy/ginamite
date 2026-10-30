package authentication

import "testing"

func TestUser_IsSaved(t *testing.T) {
	if (&User{}).IsSaved() {
		t.Error("a user without ID must not be saved")
	}

	if !(&User{ID: "abc"}).IsSaved() {
		t.Error("a user with an ID must be saved")
	}
}
