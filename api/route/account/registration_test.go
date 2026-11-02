package account

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/authentication"
)

func TestRegisterWithPassword(t *testing.T) {
	runCases(t, RegisterWithPassword, []testCase{
		{name: "invalid request", body: `{}`, setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "success", body: registrationBody(), setup: registerStubs{}.install, want: http.StatusCreated},
	})
}

func TestRegisterWithFacebook(t *testing.T) {
	runCases(t, RegisterWithFacebook, []testCase{
		{name: "invalid request", body: `{}`, setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "success", body: registrationBody(), setup: func() {
			registerStubs{}.install()
			withValidator(fakeValidator{source: authentication.MethodFacebook, ok: true})()
		}, want: http.StatusCreated},
	})
}

func TestRegisterWithGoogle(t *testing.T) {
	runCases(t, RegisterWithGoogle, []testCase{
		{name: "invalid request", body: `{}`, setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "success", body: registrationBody(), setup: func() {
			registerStubs{}.install()
			withValidator(fakeValidator{source: authentication.MethodGoogle, ok: true})()
		}, want: http.StatusCreated},
	})
}

func TestValidateRegistration(t *testing.T) {
	runCases(t, ValidateRegistration, []testCase{
		{name: "missing parameters", setup: registrationValidationStubs{}.install, want: http.StatusBadRequest},
		{name: "success", id: "u", key: "k", setup: all(registrationValidationStubs{}.install, loginStubs), want: http.StatusOK},
	})
}

func TestCancelRegistration(t *testing.T) {
	runCases(t, CancelRegistration, []testCase{
		{name: "missing parameters", setup: registrationValidationStubs{}.install, want: http.StatusBadRequest},
		{name: "success", id: "u", key: "k", setup: registrationValidationStubs{}.install, want: http.StatusOK},
	})
}
