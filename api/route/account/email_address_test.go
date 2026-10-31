package account

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUpdateEmailAddress(t *testing.T) {
	const body = `{"email_address":"new@example.com"}`
	create := func(err error) func() {
		return func() {
			testFunctions.CreateNewUpdateEmailAddressProcess = func(*gin.Context, string, string) (string, string, error) {
				return "process-id", "process-key", err
			}
		}
	}

	runCases(t, UpdateEmailAddress, []testCase{
		{name: "GetUserDetailsByID not defined", userID: testUserID, body: body, setup: create(nil), want: http.StatusInternalServerError},
		{name: "process function not defined", userID: testUserID, body: body, setup: withUser(testUser, nil), want: http.StatusInternalServerError},
		{name: "missing email address", userID: testUserID, body: `{}`, setup: all(withUser(testUser, nil), create(nil)), want: http.StatusBadRequest},
		{name: "invalid email address", userID: testUserID, body: `{"email_address":"new"}`, setup: all(withUser(testUser, nil), create(nil)), want: http.StatusBadRequest},
		{name: "user error", userID: testUserID, body: body, setup: all(withUser(testUser, errBoom), create(nil)), want: http.StatusInternalServerError},
		{name: "same email address", userID: testUserID, body: `{"email_address":"JOHN@example.com"}`, setup: all(withUser(testUser, nil), create(nil)), want: http.StatusBadRequest},
		{name: "process error", userID: testUserID, body: body, setup: all(withUser(testUser, nil), create(errBoom)), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: all(withUser(testUser, nil), create(nil)), want: http.StatusOK},
		{
			name: "success in test environment", userID: testUserID, body: body, setup: all(withUser(testUser, nil), create(nil), inTestEnvironment),
			want: http.StatusOK, check: contains("process-id", "process-key"),
		},
	})
}

func TestValidateUpdateEmailAddress(t *testing.T) {
	runCases(t, ValidateUpdateEmailAddress, []testCase{
		{name: "missing process", want: http.StatusBadRequest},
		{name: "success", id: "process", key: "secret", setup: processStubs{}.install, want: http.StatusOK},
		{name: "update error", id: "process", key: "secret", setup: processStubs{updateErr: errBoom}.install, want: http.StatusBadRequest},
	})
}

func TestCancelUpdateEmailAddress(t *testing.T) {
	runCases(t, CancelUpdateEmailAddress, []testCase{
		{name: "missing process", want: http.StatusBadRequest},
		{name: "success", id: "process", key: "secret", setup: processStubs{}.install, want: http.StatusOK},
	})
}

func TestValidateNewEmailAddress(t *testing.T) {
	runCases(t, ValidateNewEmailAddress, []testCase{
		{name: "missing process", want: http.StatusBadRequest},
		{name: "success", id: "process", key: "secret", setup: processStubs{}.install, want: http.StatusOK},
		{name: "verification error", id: "process", key: "secret", setup: processStubs{updateErr: errBoom}.install, want: http.StatusBadRequest},
	})
}

func TestConfirmEmailAddress(t *testing.T) {
	verified := testUser
	verified.IsEmailValid = true
	create := func(err error) func() {
		return func() {
			testFunctions.CreateNewVerifyEmailAddressProcess = func(*gin.Context, string) (string, string, error) {
				return "process-id", "process-key", err
			}
		}
	}

	runCases(t, ConfirmEmailAddress, []testCase{
		{name: "GetUserDetailsByID not defined", userID: testUserID, want: http.StatusInternalServerError},
		{name: "user error", userID: testUserID, setup: withUser(testUser, errBoom), want: http.StatusInternalServerError},
		{name: "already confirmed", userID: testUserID, setup: withUser(verified, nil), want: http.StatusPreconditionFailed},
		{name: "process function not defined", userID: testUserID, setup: withUser(testUser, nil), want: http.StatusInternalServerError},
		{name: "process error", userID: testUserID, setup: all(withUser(testUser, nil), create(errBoom)), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, setup: all(withUser(testUser, nil), create(nil)), want: http.StatusOK},
		{
			name: "success in test environment", userID: testUserID, setup: all(withUser(testUser, nil), create(nil), inTestEnvironment),
			want: http.StatusOK, check: contains("process-id", "process-key"),
		},
	})
}
