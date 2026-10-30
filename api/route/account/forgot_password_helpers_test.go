package account

import (
	"net/http"
	"testing"
)

// forgotPasswordStubs configures the functions used by validateOrCancelForgotPassword.
type forgotPasswordStubs struct {
	key                               string // private key of the process.
	detailsErr, userErr, updateErr    error
	deleteErr                         error
	missing                           string // name of the function to leave undefined.
	wrongAccessCalls                  *int
	skipWrongAccessFunctionDefinition bool
}

func TestValidateOrCancelForgotPassword_Validate(t *testing.T) {
	const body = `{"password":"Abcdefghij1"}`
	calls := 0

	runCases(t, ValidateForgotPassword, []testCase{
		{name: "GetUserDetailsByID not defined", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k", missing: "GetUserDetailsByID"}.install, want: http.StatusBadRequest},
		{name: "details function not defined", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k", missing: "GetForgotPasswordProcessDetailsByID"}.install, want: http.StatusBadRequest},
		{name: "UpdateUserPassword not defined", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k", missing: "UpdateUserPassword"}.install, want: http.StatusBadRequest},
		{name: "DeleteForgotPasswordProcess not defined", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k", missing: "DeleteForgotPasswordProcess"}.install, want: http.StatusBadRequest},
		{name: "missing process", body: body, setup: forgotPasswordStubs{key: "k"}.install, want: http.StatusBadRequest},
		{name: "invalid password", id: "p", key: "k", body: `{"password":"weak"}`, setup: forgotPasswordStubs{key: "k"}.install, want: http.StatusBadRequest},
		{name: "details error", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k", detailsErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "wrong key", id: "p", key: "other", body: body, setup: forgotPasswordStubs{key: "k", wrongAccessCalls: &calls}.install, want: http.StatusBadRequest},
		{name: "wrong key without handler", id: "p", key: "other", body: body, setup: forgotPasswordStubs{key: "k", skipWrongAccessFunctionDefinition: true}.install, want: http.StatusBadRequest},
		{name: "user error", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k", userErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "update error", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k", updateErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "ignores deletion error", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k", deleteErr: errBoom}.install, want: http.StatusOK},
		{name: "success", id: "p", key: "k", body: body, setup: forgotPasswordStubs{key: "k"}.install, want: http.StatusOK},
	})

	if calls != 1 {
		t.Errorf("WrongAccessToForgotPasswordProcess called %d times, want 1", calls)
	}
}

func TestValidateOrCancelForgotPassword_Cancel(t *testing.T) {
	runCases(t, CancelForgotPassword, []testCase{
		{name: "missing process", setup: forgotPasswordStubs{key: "k"}.install, want: http.StatusBadRequest},
		{name: "wrong key", id: "p", key: "other", setup: forgotPasswordStubs{key: "k"}.install, want: http.StatusBadRequest},
		{name: "deletion error", id: "p", key: "k", setup: forgotPasswordStubs{key: "k", deleteErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "success", id: "p", key: "k", setup: forgotPasswordStubs{key: "k"}.install, want: http.StatusOK},
	})
}
