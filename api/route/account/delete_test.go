package account

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func newProcess(err error) func(*gin.Context, string) (string, string, error) {
	return func(*gin.Context, string) (string, string, error) { return "process-id", "process-key", err }
}

func TestDeleteAccount(t *testing.T) {
	create := func(err error) func() {
		return func() { testFunctions.CreateNewDeleteAccountProcess = newProcess(err) }
	}

	runCases(t, DeleteAccount, []testCase{
		{name: "GetUserDetailsByID not defined", userID: testUserID, setup: create(nil), want: http.StatusInternalServerError},
		{name: "process function not defined", userID: testUserID, setup: withUser(testUser, nil), want: http.StatusInternalServerError},
		{name: "user error", userID: testUserID, setup: all(withUser(testUser, errBoom), create(nil)), want: http.StatusInternalServerError},
		{name: "process error", userID: testUserID, setup: all(withUser(testUser, nil), create(errBoom)), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, setup: all(withUser(testUser, nil), create(nil)), want: http.StatusOK},
		{
			name: "success in test environment", userID: testUserID, setup: all(withUser(testUser, nil), create(nil), inTestEnvironment),
			want: http.StatusOK, check: contains("process-id", "process-key"),
		},
	})
}

func TestValidateDeleteAccount(t *testing.T) {
	runCases(t, ValidateDeleteAccount, []testCase{
		{name: "missing process", want: http.StatusBadRequest},
		{name: "success", id: "process", key: "secret", setup: processStubs{}.install, want: http.StatusOK},
		{name: "deletion error", id: "process", key: "secret", setup: processStubs{updateErr: errBoom}.install, want: http.StatusBadRequest},
	})
}

func TestCancelDeleteAccount(t *testing.T) {
	runCases(t, CancelDeleteAccount, []testCase{
		{name: "missing process", want: http.StatusBadRequest},
		{name: "success", id: "process", key: "secret", setup: processStubs{}.install, want: http.StatusOK},
	})
}
