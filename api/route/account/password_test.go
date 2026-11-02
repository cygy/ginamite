package account

import (
	"context"
	"net/http"
	"testing"
)

func TestUpdatePassword(t *testing.T) {
	const body = `{"password":"Abcdefghij1"}`
	create := func(err error) func() {
		return func() {
			testProcesses.CreateNewUpdatePasswordProcessFn = func(context.Context, string, string) (string, string, error) {
				return "process-id", "process-key", err
			}
		}
	}

	runCases(t, UpdatePassword, []testCase{
		{name: "storage of the accounts not defined", userID: testUserID, body: body, setup: all(create(nil), withoutAccounts), want: http.StatusInternalServerError},
		{name: "storage of the processes not defined", userID: testUserID, body: body, setup: all(withUser(testUser, nil), withoutProcesses), want: http.StatusInternalServerError},
		{name: "invalid password", userID: testUserID, body: `{"password":"weak"}`, setup: all(withUser(testUser, nil), create(nil)), want: http.StatusBadRequest},
		{name: "user error", userID: testUserID, body: body, setup: all(withUser(testUser, errBoom), create(nil)), want: http.StatusInternalServerError},
		{name: "process error", userID: testUserID, body: body, setup: all(withUser(testUser, nil), create(errBoom)), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: all(withUser(testUser, nil), create(nil)), want: http.StatusOK},
		{
			name: "success in test environment", userID: testUserID, body: body, setup: all(withUser(testUser, nil), create(nil), inTestEnvironment),
			want: http.StatusOK, check: contains("process-id", "process-key"),
		},
	})
}

func TestValidateUpdatePassword(t *testing.T) {
	runCases(t, ValidateUpdatePassword, []testCase{
		{name: "missing process", want: http.StatusBadRequest},
		{name: "success", id: "process", key: "secret", setup: processStubs{}.install, want: http.StatusOK},
		{name: "update error", id: "process", key: "secret", setup: processStubs{updateErr: errBoom}.install, want: http.StatusBadRequest},
	})
}

func TestCancelUpdatePassword(t *testing.T) {
	runCases(t, CancelUpdatePassword, []testCase{
		{name: "missing process", want: http.StatusBadRequest},
		{name: "success", id: "process", key: "secret", setup: processStubs{}.install, want: http.StatusOK},
	})
}
