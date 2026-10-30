package contact

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/common"

	"github.com/gin-gonic/gin"
)

func setup(t *testing.T) {
	t.Helper()

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})
	t.Cleanup(func() {
		CreateContact, UpdateContact, GetContacts, GetContact = nil, nil, nil, nil
	})
}

func serve(handler gin.HandlerFunc, method, target, body string, middlewares ...gin.HandlerFunc) int {
	router := gin.New()
	router.Use(ctxLocale.Inject("en", []string{"en"}))
	router.Use(middlewares...)
	router.Any("/:id", handler)
	router.Any("/", handler)

	req, _ := http.NewRequest(method, target, strings.NewReader(body))

	return apitest.Serve(router, req).Code
}

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
			CreateContact = tt.create
			if got := serve(Post, http.MethodPost, "/", tt.body); got != tt.wantStatus {
				t.Errorf("status = %d, want %d", got, tt.wantStatus)
			}
		})
	}

	t.Run("created and stored", func(t *testing.T) {
		var email, subject, text string
		CreateContact = func(_ *gin.Context, e, s, x string) error {
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

func TestGetAdminContacts(t *testing.T) {
	setup(t)

	if got := serve(GetAdminContacts, http.MethodGet, "/", ""); got != http.StatusInternalServerError {
		t.Errorf("undefined function: status = %d, want %d", got, http.StatusInternalServerError)
	}

	var gotOffset, gotLimit int64
	GetContacts = func(_ *gin.Context, offset, limit int64) (any, error) {
		gotOffset, gotLimit = offset, limit
		return []string{"c"}, nil
	}
	if got := serve(GetAdminContacts, http.MethodGet, "/?offset=2&limit=5", "", request.VerifyLimitAndOffset(0, 10, 50)); got != http.StatusOK {
		t.Fatalf("status = %d, want %d", got, http.StatusOK)
	}
	if gotOffset != 2 || gotLimit != 5 {
		t.Errorf("offset, limit = %d, %d, want 2, 5", gotOffset, gotLimit)
	}

	GetContacts = func(*gin.Context, int64, int64) (any, error) { return nil, errors.New("boom") }
	if got := serve(GetAdminContacts, http.MethodGet, "/", ""); got != http.StatusInternalServerError {
		t.Errorf("failing function: status = %d, want %d", got, http.StatusInternalServerError)
	}
}

func TestGetAdminContact(t *testing.T) {
	setup(t)

	if got := serve(GetAdminContact, http.MethodGet, "/1", ""); got != http.StatusInternalServerError {
		t.Errorf("undefined function: status = %d, want %d", got, http.StatusInternalServerError)
	}

	var gotID string
	GetContact = func(_ *gin.Context, id string) (any, error) {
		gotID = id
		return "c", nil
	}
	if got := serve(GetAdminContact, http.MethodGet, "/abc", ""); got != http.StatusOK || gotID != "abc" {
		t.Errorf("status = %d, id = %q, want 200 and abc", got, gotID)
	}

	GetContact = func(*gin.Context, string) (any, error) { return nil, errors.New("boom") }
	if got := serve(GetAdminContact, http.MethodGet, "/abc", ""); got != http.StatusInternalServerError {
		t.Errorf("failing function: status = %d, want %d", got, http.StatusInternalServerError)
	}
}

func TestUpdate(t *testing.T) {
	setup(t)

	if got := serve(Update, http.MethodPut, "/1", `{"done":true}`); got != http.StatusInternalServerError {
		t.Errorf("undefined function: status = %d, want %d", got, http.StatusInternalServerError)
	}

	var gotID string
	var gotDone bool
	UpdateContact = func(_ *gin.Context, id string, done bool) error {
		gotID, gotDone = id, done
		return nil
	}
	if got := serve(Update, http.MethodPut, "/abc", `{"done":true}`); got != http.StatusOK || gotID != "abc" || !gotDone {
		t.Errorf("status = %d, id = %q, done = %v, want 200, abc, true", got, gotID, gotDone)
	}

	UpdateContact = func(*gin.Context, string, bool) error { return errors.New("boom") }
	if got := serve(Update, http.MethodPut, "/abc", `{"done":true}`); got != http.StatusInternalServerError {
		t.Errorf("failing function: status = %d, want %d", got, http.StatusInternalServerError)
	}
}
