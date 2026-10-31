package contact

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common"

	"github.com/gin-gonic/gin"
)

func TestPost(t *testing.T) {
	setup(t)

	long := func(n int) string { return strings.Repeat("a", n) }
	body := func(email, subject, text string) string {
		return `{"email_address":"` + email + `","subject":"` + subject + `","text":"` + text + `"}`
	}

	tests := []struct {
		name       string
		body       string
		create     CreateContactFunc
		wantStatus int
	}{
		{"missing email address", body("", "s", "t"), nil, http.StatusBadRequest},
		{"missing subject", body("a@b.c", "", "t"), nil, http.StatusBadRequest},
		{"missing text", body("a@b.c", "s", ""), nil, http.StatusBadRequest},
		{"subject too long", body("a@b.c", long(common.ContactSubjectMaxLength+1), "t"), nil, http.StatusBadRequest},
		{"text too long", body("a@b.c", "s", long(common.ContactTextMaxLength+1)), nil, http.StatusBadRequest},
		{"invalid email address", body("abc", "s", "t"), nil, http.StatusBadRequest},
		{"created without storage", body("a@b.c", "s", "t"), nil, http.StatusCreated},
		{"storage fails", body("a@b.c", "s", "t"), func(*gin.Context, string, string, string) error { return errors.New("boom") }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFunctions.CreateContact = tt.create
			if got := serve(Post, http.MethodPost, "/", tt.body); got != tt.wantStatus {
				t.Errorf("status = %d, want %d", got, tt.wantStatus)
			}
		})
	}

	t.Run("created and stored", func(t *testing.T) {
		var email, subject, text string
		testFunctions.CreateContact = func(_ *gin.Context, e, s, x string) error {
			email, subject, text = e, s, x
			return nil
		}

		if got := serve(Post, http.MethodPost, "/", body("a@b.c", "subject", "text")); got != http.StatusCreated {
			t.Fatalf("status = %d, want %d", got, http.StatusCreated)
		}
		if email != "a@b.c" || subject != "subject" || text != "text" {
			t.Errorf("stored = %q, %q, %q", email, subject, text)
		}
	})
}
