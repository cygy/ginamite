package account

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"

	"github.com/gin-gonic/gin"
)

func TestForgotPassword(t *testing.T) {
	const body = `{"identifier":"john"}`
	lookup := func(err error) func() {
		return func() {
			testFunctions.GetUserDetailsByIdentifier = func(*gin.Context, string) (authentication.User, error) { return testUser, err }
		}
	}
	create := func(err error) func() {
		return func() {
			testFunctions.NewForgotPasswordProcessForUser = func(*gin.Context, string, string) (string, error) { return "process-id", err }
		}
	}

	runCases(t, ForgotPassword, []testCase{
		{name: "lookup function not defined", body: body, setup: create(nil), want: http.StatusInternalServerError},
		{name: "process function not defined", body: body, setup: lookup(nil), want: http.StatusInternalServerError},
		{name: "missing identifier", body: `{}`, setup: all(lookup(nil), create(nil)), want: http.StatusBadRequest},
		{name: "unknown account", body: body, setup: all(lookup(commonErrors.NotFound()), create(nil)), want: http.StatusNotFound},
		{name: "lookup error", body: body, setup: all(lookup(errBoom), create(nil)), want: http.StatusInternalServerError},
		{name: "process error", body: body, setup: all(lookup(nil), create(errBoom)), want: http.StatusInternalServerError},
		{name: "success", body: body, setup: all(lookup(nil), create(nil)), want: http.StatusOK},
		{
			name: "success in test environment", body: body, setup: all(lookup(nil), create(nil), inTestEnvironment),
			want: http.StatusOK, check: contains("process-id", "key"),
		},
	})
}

func (f forgotPasswordStubs) install() {
	testFunctions.GetForgotPasswordProcessDetailsByID = func(*gin.Context, string) (string, string, error) {
		return testUserID, f.key, f.detailsErr
	}
	testFunctions.GetUserDetailsByID = func(*gin.Context, string) (authentication.User, error) { return testUser, f.userErr }
	testFunctions.UpdateUserPassword = func(*gin.Context, string, string, bool) error { return f.updateErr }
	testFunctions.DeleteForgotPasswordProcess = func(*gin.Context, string) error { return f.deleteErr }
	if !f.skipWrongAccessFunctionDefinition {
		testFunctions.WrongAccessToForgotPasswordProcess = func(*gin.Context, string) {
			if f.wrongAccessCalls != nil {
				*f.wrongAccessCalls++
			}
		}
	}

	switch f.missing {
	case "GetUserDetailsByID":
		testFunctions.GetUserDetailsByID = nil
	case "GetForgotPasswordProcessDetailsByID":
		testFunctions.GetForgotPasswordProcessDetailsByID = nil
	case "UpdateUserPassword":
		testFunctions.UpdateUserPassword = nil
	case "DeleteForgotPasswordProcess":
		testFunctions.DeleteForgotPasswordProcess = nil
	}
}
