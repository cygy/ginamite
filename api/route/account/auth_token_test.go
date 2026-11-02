package account

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"

	"github.com/gin-gonic/gin"
	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

const loginDevice = `"source":"web","device":{"type":"mobile","name":"iPhone"}`

// loginStubs defines the functions needed to create an authentication token.
func loginStubs() {
	testTokens.saveFn = func(context.Context, authentication.Details) (string, bool, error) { return testTokenID, false, nil }
	testTokens.deleteFn = func(context.Context, string) error { return nil }
}

func TestLoginWithPassword(t *testing.T) {
	encrypted, err := authentication.EncryptPassword("Abcdefghij1")
	if err != nil {
		t.Fatalf("unable to encrypt the password: %v", err)
	}
	lookup := func(password string, err error) func() {
		return func() {
			testFunctions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser = func(*gin.Context, string) (authentication.User, string, error) {
				return testUser, password, err
			}
		}
	}
	body := `{"identifier":"john","password":"Abcdefghij1",` + loginDevice + `}`

	runCases(t, LoginWithPassword, []testCase{
		{name: "function not defined", body: body, want: http.StatusInternalServerError},
		{name: "missing identifier", body: `{"password":"x",` + loginDevice + `}`, setup: lookup(encrypted, nil), want: http.StatusBadRequest},
		{name: "missing password", body: `{"identifier":"john",` + loginDevice + `}`, setup: lookup(encrypted, nil), want: http.StatusBadRequest},
		{name: "missing source", body: `{"identifier":"john","password":"x","device":{"type":"mobile","name":"iPhone"}}`, setup: lookup(encrypted, nil), want: http.StatusBadRequest},
		{name: "missing device", body: `{"identifier":"john","password":"x","source":"web"}`, setup: lookup(encrypted, nil), want: http.StatusBadRequest},
		{name: "lookup error", body: body, setup: lookup("", errBoom), want: http.StatusBadRequest},
		{name: "wrong password", body: `{"identifier":"john","password":"Wrongpass1",` + loginDevice + `}`, setup: lookup(encrypted, nil), want: http.StatusBadRequest},
		{name: "token creation failure", body: body, setup: all(lookup(encrypted, nil), withoutTokens), want: http.StatusBadRequest},
		{name: "success", body: body, setup: all(lookup(encrypted, nil), loginStubs), want: http.StatusCreated, check: contains("auth_token")},
	})
}

func TestLogout(t *testing.T) {
	remove := func(err error) func() {
		return func() { testTokens.deleteFn = func(context.Context, string) error { return err } }
	}
	expiration := func(err error) func() {
		return func() {
			testTokens.expirationFn = func(context.Context, string) (time.Time, error) { return time.Now().Add(time.Hour), err }
		}
	}

	runCases(t, Logout, []testCase{
		{name: "storage not defined", setup: withoutTokens, userID: testUserID, want: http.StatusInternalServerError},
		{name: "unknown token", userID: testUserID, setup: remove(commonErrors.NotFound()), want: http.StatusNoContent},
		{name: "deletion error", userID: testUserID, setup: remove(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, setup: remove(nil), want: http.StatusNoContent},
		{name: "success with expiration date", userID: testUserID, setup: all(remove(nil), expiration(nil)), want: http.StatusNoContent},
		{name: "success with expiration date error", userID: testUserID, setup: all(remove(nil), expiration(errBoom)), want: http.StatusNoContent},
	})
}

func TestGetAuthTokens(t *testing.T) {
	list := func(err error) func() {
		return func() {
			testTokens.ownedFn = func(context.Context, string) (any, error) { return []string{"t1"}, err }
		}
	}

	runCases(t, GetAuthTokens, []testCase{
		{name: "storage not defined", setup: withoutTokens, userID: testUserID, want: http.StatusInternalServerError},
		{name: "not found", userID: testUserID, setup: list(commonErrors.NotFound()), want: http.StatusNotFound},
		{name: "error", userID: testUserID, setup: list(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, setup: list(nil), want: http.StatusOK, check: contains("t1")},
	})
}

func TestGetAuthToken(t *testing.T) {
	get := func(owner string, err error) func() {
		return func() {
			testTokens.ownedOneFn = func(context.Context, string) (any, string, error) { return "t1", owner, err }
		}
	}

	runCases(t, GetAuthToken, []testCase{
		{name: "storage not defined", setup: withoutTokens, userID: testUserID, id: "t1", want: http.StatusInternalServerError},
		{name: "not found", userID: testUserID, id: "t1", setup: get("", commonErrors.NotFound()), want: http.StatusNotFound},
		{name: "error", userID: testUserID, id: "t1", setup: get("", errBoom), want: http.StatusInternalServerError},
		{name: "token of another user", userID: testUserID, id: "t1", setup: get("someone-else", nil), want: http.StatusNotFound},
		{name: "success", userID: testUserID, id: "t1", setup: get(testUserID, nil), want: http.StatusOK, check: contains("t1")},
	})
}

func TestDeleteAuthToken(t *testing.T) {
	token := authentication.Token{UserID: testUserID, PrivateKey: "token-key"}

	runCases(t, DeleteAuthToken, []testCase{
		{name: "token of another user", userID: "someone-else", id: "t1", setup: deleteStubs(token, nil, nil), want: http.StatusNotFound},
		{name: "success", userID: testUserID, id: "t1", setup: deleteStubs(token, nil, nil), want: http.StatusOK},
	})
}

func TestDeleteAuthTokenWithKey(t *testing.T) {
	token := authentication.Token{UserID: testUserID, PrivateKey: "token-key"}

	runCases(t, DeleteAuthTokenWithKey, []testCase{
		{name: "missing key", id: "t1", setup: deleteStubs(token, nil, nil), want: http.StatusBadRequest},
		{name: "wrong key", id: "t1", key: "other", setup: deleteStubs(token, nil, nil), want: http.StatusBadRequest},
		{name: "success", id: "t1", key: "token-key", setup: deleteStubs(token, nil, nil), want: http.StatusOK},
	})
}

func TestUpdateAuthToken(t *testing.T) {
	const body = `{"name":"My phone","notifications":["new_login"]}`
	token := authentication.Token{UserID: testUserID}
	stubs := func(owner authentication.Token, getErr, updateErr error) func() {
		return func() {
			testTokens.detailsFn = func(context.Context, string) (authentication.Token, error) { return owner, getErr }
			testTokens.updateFn = func(context.Context, string, string, bool, []string) error { return updateErr }
		}
	}

	runCases(t, UpdateAuthToken, []testCase{
		{name: "storage not defined", userID: testUserID, id: "t1", body: body, setup: withoutTokens, want: http.StatusInternalServerError},
		{name: "missing name", userID: testUserID, id: "t1", body: `{}`, setup: stubs(token, nil, nil), want: http.StatusBadRequest},
		{name: "unsupported notification", userID: testUserID, id: "t1", body: `{"name":"x","notifications":["unknown"]}`, setup: stubs(token, nil, nil), want: http.StatusBadRequest},
		{name: "token not found", userID: testUserID, id: "t1", body: body, setup: stubs(token, commonErrors.NotFound(), nil), want: http.StatusNotFound},
		{name: "token of another user", userID: "someone-else", id: "t1", body: body, setup: stubs(token, nil, nil), want: http.StatusNotFound},
		{name: "update error", userID: testUserID, id: "t1", body: body, setup: stubs(token, nil, errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, id: "t1", body: body, setup: stubs(token, nil, nil), want: http.StatusOK},
	})
}

func TestRefreshAuthToken(t *testing.T) {
	runCases(t, RefreshAuthToken, []testCase{
		{name: "success", userID: testUserID, want: http.StatusOK, check: contains("auth_token")},
		{name: "refresh error", userID: testUserID, setup: func() {
			testHooks.ExtendTokenExpirationDateFromID = func(context.Context, string, time.Duration) error { return errBoom }
		}, want: http.StatusBadRequest},
	})
}

func TestLoginWithPassword_UpgradesTheLegacyHash(t *testing.T) {
	legacy, err := sha512_crypt.New().Generate([]byte("Abcdefghij1"), []byte("$6$rounds=5000$abcdefghijklmnop$"))
	if err != nil {
		t.Fatal(err)
	}
	current, err := authentication.EncryptPassword("Abcdefghij1")
	if err != nil {
		t.Fatal(err)
	}
	body := `{"identifier":"john","password":"Abcdefghij1",` + loginDevice + `}`

	run := func(stored string, updateErr error) (updated string, calls int) {
		lookup := func() {
			testFunctions.GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser = func(*gin.Context, string) (authentication.User, string, error) {
				return testUser, stored, nil
			}
			testFunctions.UpdateEncryptedPasswordByUserID = func(_ *gin.Context, userID, encrypted string) error {
				if userID != testUser.ID {
					t.Errorf("user = %q, want %q", userID, testUser.ID)
				}
				updated = encrypted
				calls++

				return updateErr
			}
		}
		runCases(t, LoginWithPassword, []testCase{{name: "login", body: body, setup: all(lookup, loginStubs), want: http.StatusCreated, check: contains("auth_token")}})

		return updated, calls
	}

	t.Run("a legacy hash is replaced", func(t *testing.T) {
		updated, calls := run(legacy, nil)
		if calls != 1 || !strings.HasPrefix(updated, "$argon2id$") || !authentication.ComparePassword("Abcdefghij1", updated) {
			t.Errorf("calls = %d, updated = %q", calls, updated)
		}
	})

	t.Run("a current hash is kept", func(t *testing.T) {
		if _, calls := run(current, nil); calls != 0 {
			t.Errorf("calls = %d, want 0", calls)
		}
	})

	t.Run("a failure does not prevent the login", func(t *testing.T) {
		if _, calls := run(legacy, errBoom); calls != 1 {
			t.Errorf("calls = %d, want 1", calls)
		}
	})
}
