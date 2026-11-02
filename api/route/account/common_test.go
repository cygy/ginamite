package account

import (
	"context"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// processStubs configures the functions used by validateOrCancelUpdateProperty.
type processStubs struct {
	verifyErr, userErr, updateErr, deleteProcessErr error
	missing                                         string // name of the function to leave undefined.
}

func (p processStubs) install() {
	testProcesses.VerifyProcessFn = func(context.Context, string, string) (string, string, error) { return testUserID, "value", p.verifyErr }
	testAccounts.GetUserDetailsByIDFn = func(context.Context, string) (authentication.User, error) { return testUser, p.userErr }
	testProcesses.DeleteProcessByIDFn = func(context.Context, string) error { return p.deleteProcessErr }
	testAccounts.UpdateUserEmailAddressFn = func(context.Context, string, string) error { return p.updateErr }
	testAccounts.UpdateUserPasswordFn = func(context.Context, string, string, bool) error { return p.updateErr }
	testAccounts.UserEmailAddressIsVerifiedFn = func(context.Context, string) error { return p.updateErr }
	testAccounts.DeleteUserByIDFn = func(context.Context, string) error { return p.updateErr }
	testAccounts.DisableUserByIDFn = func(context.Context, string) (authentication.User, error) { return testUser, p.updateErr }
	testProcesses.CreateNewVerifyEmailAddressProcessFn = func(context.Context, string) (string, string, error) { return "id", "key", nil }

	switch p.missing {
	case "Accounts":
		withoutAccounts()
	case "Processes":
		withoutProcesses()
	}
}

func TestValidateOrCancelUpdateProperty(t *testing.T) {
	properties := []string{
		authentication.ProcessUpdateEmailAddress,
		authentication.ProcessUpdatePassword,
		authentication.ProcessVerifyEmailAddress,
		authentication.ProcessDeleteAccount,
		authentication.ProcessDisableAccount,
	}
	params := gin.Params{{Key: "id", Value: "process"}, {Key: "key", Value: "secret"}}

	tests := []struct {
		name     string
		validate bool
		stubs    processStubs
		params   gin.Params
		want     int
	}{
		{"validate success", true, processStubs{}, params, http.StatusOK},
		{"validate missing key", true, processStubs{}, params[:1], http.StatusBadRequest},
		{"validate missing process", true, processStubs{}, nil, http.StatusBadRequest},
		{"validate without the accounts", true, processStubs{missing: "Accounts"}, params, http.StatusBadRequest},
		{"validate without the processes", true, processStubs{missing: "Processes"}, params, http.StatusBadRequest},
		{"validate verify error", true, processStubs{verifyErr: errBoom}, params, http.StatusBadRequest},
		{"validate user error", true, processStubs{userErr: errBoom}, params, http.StatusBadRequest},
		{"validate update error", true, processStubs{updateErr: errBoom}, params, http.StatusBadRequest},
		{"validate ignores process deletion error", true, processStubs{deleteProcessErr: errBoom}, params, http.StatusOK},
		{"cancel success", false, processStubs{}, params, http.StatusOK},
		{"cancel without the accounts", false, processStubs{missing: "Accounts"}, params, http.StatusBadRequest},
		{"cancel verify error", false, processStubs{verifyErr: errBoom}, params, http.StatusBadRequest},
		{"cancel process deletion error", false, processStubs{deleteProcessErr: errBoom}, params, http.StatusBadRequest},
	}
	for _, property := range properties {
		for _, tt := range tests {
			t.Run(property+"/"+tt.name, func(t *testing.T) {
				setup(t)
				tt.stubs.install()

				c, recorder := newContext(tt.params)
				validateOrCancelUpdateProperty(c, property, tt.validate)
				c.Writer.WriteHeaderNow()

				if recorder.Code != tt.want {
					t.Errorf("status = %d, want %d", recorder.Code, tt.want)
				}
			})
		}
	}
}

func TestValidateOrCancelUpdateProperty_WrongProcess(t *testing.T) {
	setup(t)

	c, recorder := newContext(nil)
	validateOrCancelUpdateProperty(c, "unknown", true)
	c.Writer.WriteHeaderNow()

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
		status   int
	}{
		{"empty", "", false, http.StatusBadRequest},
		{"too short", "Ab1", false, http.StatusBadRequest},
		{"no lowercase", "ABCDEFGHIJ1", false, http.StatusBadRequest},
		{"no uppercase", "abcdefghij1", false, http.StatusBadRequest},
		{"no digit", "Abcdefghijk", false, http.StatusBadRequest},
		{"valid", "Abcdefghij1", true, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)

			c, recorder := newContext(nil)
			if got := validatePassword(c, tt.password, "placeholder"); got != tt.want {
				t.Fatalf("validatePassword() = %v, want %v", got, tt.want)
			}
			c.Writer.WriteHeaderNow()
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d", recorder.Code, tt.status)
			}
		})
	}
}

func TestValidateEmailAddress(t *testing.T) {
	tests := []struct {
		name         string
		emailAddress string
		want         bool
		status       int
	}{
		{"empty", "", false, http.StatusBadRequest},
		{"no at sign", "john.example.com", false, http.StatusBadRequest},
		{"valid", "john@example.com", true, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)

			c, recorder := newContext(nil)
			if got := validateEmailAddress(c, tt.emailAddress, "placeholder"); got != tt.want {
				t.Fatalf("validateEmailAddress() = %v, want %v", got, tt.want)
			}
			c.Writer.WriteHeaderNow()
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d", recorder.Code, tt.status)
			}
		})
	}
}
