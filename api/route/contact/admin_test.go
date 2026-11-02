package contact

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/api/request"
)

func TestGetAdminContacts(t *testing.T) {
	setup(t)

	if got := serve(GetAdminContacts, http.MethodGet, "/", "", withoutStorage); got != http.StatusInternalServerError {
		t.Errorf("no storage: status = %d, want %d", got, http.StatusInternalServerError)
	}

	var gotOffset, gotLimit int64
	testStorage.list = func(_ context.Context, offset, limit int64) (any, error) {
		gotOffset, gotLimit = offset, limit
		return []string{"c"}, nil
	}
	if got := serve(GetAdminContacts, http.MethodGet, "/?offset=2&limit=5", "", request.VerifyLimitAndOffset(0, 10, 50)); got != http.StatusOK {
		t.Fatalf("status = %d, want %d", got, http.StatusOK)
	}
	if gotOffset != 2 || gotLimit != 5 {
		t.Errorf("offset, limit = %d, %d, want 2, 5", gotOffset, gotLimit)
	}

	testStorage.list = func(context.Context, int64, int64) (any, error) { return nil, errors.New("boom") }
	if got := serve(GetAdminContacts, http.MethodGet, "/", "", withoutStorage); got != http.StatusInternalServerError {
		t.Errorf("failing function: status = %d, want %d", got, http.StatusInternalServerError)
	}
}

func TestGetAdminContact(t *testing.T) {
	setup(t)

	if got := serve(GetAdminContact, http.MethodGet, "/1", "", withoutStorage); got != http.StatusInternalServerError {
		t.Errorf("no storage: status = %d, want %d", got, http.StatusInternalServerError)
	}

	var gotID string
	testStorage.get = func(_ context.Context, id string) (any, error) {
		gotID = id
		return "c", nil
	}
	if got := serve(GetAdminContact, http.MethodGet, "/abc", ""); got != http.StatusOK || gotID != "abc" {
		t.Errorf("status = %d, id = %q, want 200 and abc", got, gotID)
	}

	testStorage.get = func(context.Context, string) (any, error) { return nil, errors.New("boom") }
	if got := serve(GetAdminContact, http.MethodGet, "/abc", ""); got != http.StatusInternalServerError {
		t.Errorf("failing function: status = %d, want %d", got, http.StatusInternalServerError)
	}
}

func TestUpdate(t *testing.T) {
	setup(t)

	if got := serve(Update, http.MethodPut, "/1", `{"done":true}`, withoutStorage); got != http.StatusInternalServerError {
		t.Errorf("no storage: status = %d, want %d", got, http.StatusInternalServerError)
	}

	var gotID string
	var gotDone bool
	testStorage.update = func(_ context.Context, id string, done bool) error {
		gotID, gotDone = id, done
		return nil
	}
	if got := serve(Update, http.MethodPut, "/abc", `{"done":true}`); got != http.StatusOK || gotID != "abc" || !gotDone {
		t.Errorf("status = %d, id = %q, done = %v, want 200, abc, true", got, gotID, gotDone)
	}

	testStorage.update = func(context.Context, string, bool) error { return errors.New("boom") }
	if got := serve(Update, http.MethodPut, "/abc", `{"done":true}`); got != http.StatusInternalServerError {
		t.Errorf("failing function: status = %d, want %d", got, http.StatusInternalServerError)
	}
}
