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
			GetUserDetailsByIdentifier = func(*gin.Context, string) (authentication.User, error) { return testUser, err }
		}
	}
	create := func(err error) func() {
		return func() {
			NewForgotPasswordProcessForUser = func(*gin.Context, string, string) (string, error) { return "process-id", err }
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
	GetForgotPasswordProcessDetailsByID = func(*gin.Context, string) (string, string, error) {
		return testUserID, f.key, f.detailsErr
	}
	GetUserDetailsByID = func(*gin.Context, string) (authentication.User, error) { return testUser, f.userErr }
	UpdateUserPassword = func(*gin.Context, string, string, bool) error { return f.updateErr }
	DeleteForgotPasswordProcess = func(*gin.Context, string) error { return f.deleteErr }
	if !f.skipWrongAccessFunctionDefinition {
		WrongAccessToForgotPasswordProcess = func(*gin.Context, string) {
			if f.wrongAccessCalls != nil {
				*f.wrongAccessCalls++
			}
		}
	}

	switch f.missing {
	case "GetUserDetailsByID":
		GetUserDetailsByID = nil
	case "GetForgotPasswordProcessDetailsByID":
		GetForgotPasswordProcessDetailsByID = nil
	case "UpdateUserPassword":
		UpdateUserPassword = nil
	case "DeleteForgotPasswordProcess":
		DeleteForgotPasswordProcess = nil
	}
}
