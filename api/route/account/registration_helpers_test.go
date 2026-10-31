package account

import (
	"net/http"
	"strings"
	"testing"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"

	"github.com/gin-gonic/gin"
)

// registrationBody returns a valid registration body, the fields being overridden by the replacements.
func registrationBody(replacements ...string) string {
	body := `{"username":"Cyril","email_address":"john@example.com","password":"Abcdefghij1","token":"third-party-token",` +
		`"source":"web","device":{"type":"mobile","name":"iPhone"},"locale":"en","terms_version":"v1"}`
	for i := 0; i < len(replacements); i += 2 {
		body = strings.Replace(body, replacements[i], replacements[i+1], 1)
	}

	return body
}

// registerStubs configures the functions used by register.
type registerStubs struct {
	usernameExists, emailExists, socialExists bool
	usernameErr, emailErr, socialErr          error
	registerErr, identifierErr                error
	missing                                   string // name of the function to leave undefined.
}

func (r registerStubs) install() {
	testFunctions.DoesAccountWithUsernameExist = func(*gin.Context, string) (bool, error) { return r.usernameExists, r.usernameErr }
	testFunctions.DoesAccountWithEmailAddressExist = func(*gin.Context, string) (bool, error) { return r.emailExists, r.emailErr }
	testFunctions.DoesUserWithFacebookUserIDExist = func(*gin.Context, string) (bool, error) { return r.socialExists, r.socialErr }
	testFunctions.DoesUserWithGoogleUserIDExist = func(*gin.Context, string) (bool, error) { return r.socialExists, r.socialErr }
	testFunctions.GetUserDetailsByIdentifier = func(*gin.Context, string) (authentication.User, error) { return testUser, r.identifierErr }
	testFunctions.RegisterByEmailAddress = func(*gin.Context, string, string, string, string, string, string, string, string, string, authentication.Device) error {
		return r.registerErr
	}
	testFunctions.RegisterByThirdPartyToken = func(*gin.Context, string, validator.TokenInfos, string, string, string, string, string, string, string, authentication.Device) error {
		return r.registerErr
	}

	switch r.missing {
	case "GetUserDetailsByIdentifier":
		testFunctions.GetUserDetailsByIdentifier = nil
	case "RegisterByEmailAddress":
		testFunctions.RegisterByEmailAddress = nil
	case "RegisterByThirdPartyToken":
		testFunctions.RegisterByThirdPartyToken = nil
	case "DoesAccountWithUsernameExist":
		testFunctions.DoesAccountWithUsernameExist = nil
	case "DoesAccountWithEmailAddressExist":
		testFunctions.DoesAccountWithEmailAddressExist = nil
	}
}

func TestRegister_Password(t *testing.T) {
	handler := func(c *gin.Context) { register(c, authentication.MethodPassword) }
	body := registrationBody()
	notConfirmed := func() { testConfig.Account.EmailAddressMustBeConfirmed = false }

	missing := func(name string) testCase {
		return testCase{name: name + " not defined", body: body, setup: registerStubs{missing: name}.install, want: http.StatusInternalServerError}
	}

	runCases(t, handler, []testCase{
		missing("GetUserDetailsByIdentifier"),
		missing("RegisterByEmailAddress"),
		missing("RegisterByThirdPartyToken"),
		missing("DoesAccountWithUsernameExist"),
		missing("DoesAccountWithEmailAddressExist"),
		{name: "invalid username", body: registrationBody("Cyril", "ab"), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "invalid email address", body: registrationBody("john@example.com", "john"), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "invalid password", body: registrationBody("Abcdefghij1", "weak"), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "missing source", body: registrationBody(`"source":"web",`, ""), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "missing device", body: registrationBody(`"device":{"type":"mobile","name":"iPhone"},`, ""), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "missing terms version", body: registrationBody(`,"terms_version":"v1"`, ""), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "wrong terms version", body: registrationBody(`"v1"`, `"v0"`), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "missing locale", body: registrationBody(`"locale":"en",`, ""), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "unsupported locale", body: registrationBody(`"locale":"en"`, `"locale":"xx"`), setup: registerStubs{}.install, want: http.StatusBadRequest},
		{name: "username taken", body: body, setup: registerStubs{usernameExists: true}.install, want: http.StatusPreconditionFailed},
		{name: "username lookup error", body: body, setup: registerStubs{usernameErr: errBoom}.install, want: http.StatusPreconditionFailed},
		{name: "email address taken", body: body, setup: registerStubs{emailExists: true}.install, want: http.StatusPreconditionFailed},
		{name: "email address lookup error", body: body, setup: registerStubs{emailErr: errBoom}.install, want: http.StatusPreconditionFailed},
		{name: "registration error", body: body, setup: registerStubs{registerErr: errBoom}.install, want: http.StatusInternalServerError},
		{name: "user lookup error", body: body, setup: registerStubs{identifierErr: errBoom}.install, want: http.StatusInternalServerError},
		{name: "success", body: body, setup: registerStubs{}.install, want: http.StatusCreated, check: contains("status")},
		{
			name: "success in test environment", body: body, setup: all(registerStubs{}.install, inTestEnvironment),
			want: http.StatusCreated, check: contains(`"key"`, `"id"`),
		},
		{
			name: "success with automatic login", body: body, setup: all(registerStubs{}.install, notConfirmed, loginStubs),
			want: http.StatusCreated, check: contains("auth_token"),
		},
	})
}

func TestRegister_ThirdParty(t *testing.T) {
	infos := validator.TokenInfos{UserID: "social-id", Email: "john@example.com"}

	for _, method := range []string{authentication.MethodFacebook, authentication.MethodGoogle} {
		handler := func(c *gin.Context) { register(c, method) }
		valid := withValidator(fakeValidator{source: method, infos: infos, ok: true})
		body := registrationBody()

		t.Run(method, func(t *testing.T) {
			runCases(t, handler, []testCase{
				{name: "missing token", body: registrationBody(`"token":"third-party-token",`, ""), setup: registerStubs{}.install, want: http.StatusBadRequest},
				{name: "invalid third-party token", body: body, setup: all(registerStubs{}.install, withValidator(fakeValidator{source: method})), want: http.StatusPreconditionFailed},
				{name: "email address taken", body: body, setup: all(registerStubs{emailExists: true}.install, valid), want: http.StatusPreconditionFailed},
				{name: "registration error", body: body, setup: all(registerStubs{registerErr: errBoom}.install, valid), want: http.StatusInternalServerError},
				{name: "user lookup error", body: body, setup: all(registerStubs{identifierErr: errBoom}.install, valid), want: http.StatusInternalServerError},
				{name: "success", body: body, setup: all(registerStubs{}.install, valid), want: http.StatusCreated, check: contains("status")},
				{name: "success with automatic login", body: body, setup: all(registerStubs{}.install, valid, func() {
					testConfig.Account.EmailAddressMustBeConfirmed = false
				}, loginStubs), want: http.StatusCreated, check: contains("auth_token")},
				{
					name: "social account already registered", body: body, setup: all(registerStubs{socialExists: true}.install, valid),
					want: http.StatusOK,
				},
				{
					name: "social account lookup error", body: body, setup: all(registerStubs{socialErr: errBoom}.install, valid),
					want: http.StatusOK,
				},
				{name: "social lookup function not defined", body: body, setup: all(registerStubs{}.install, valid, func() {
					testFunctions.DoesUserWithFacebookUserIDExist, testFunctions.DoesUserWithGoogleUserIDExist = nil, nil
				}), want: http.StatusCreated},
			})
		})
	}
}

func TestRegister_WrongMethod(t *testing.T) {
	setup(t)
	registerStubs{}.install()

	c, recorder := newContext(nil)
	register(c, "unknown")
	c.Writer.WriteHeaderNow()

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

// registrationValidationStubs configures the functions used by validateOrCancelRegistration.
type registrationValidationStubs struct {
	userErr, detailsErr, validateErr, cancelErr error
	missing                                     string // name of the function to leave undefined.
}

func (r registrationValidationStubs) install() {
	testFunctions.GetUserDetailsByID = func(*gin.Context, string) (authentication.User, error) { return testUser, r.userErr }
	testFunctions.GetRegistrationDetails = func(*gin.Context, string) (*authentication.Details, error) {
		return &authentication.Details{Source: "web", Method: authentication.MethodPassword, Device: authentication.Device{Type: "mobile", Name: "iPhone"}}, r.detailsErr
	}
	testFunctions.ValidateUserRegistration = func(*gin.Context, string, string) error { return r.validateErr }
	testFunctions.CancelUserRegistration = func(*gin.Context, string, string) error { return r.cancelErr }

	switch r.missing {
	case "GetUserDetailsByID":
		testFunctions.GetUserDetailsByID = nil
	case "GetRegistrationDetails":
		testFunctions.GetRegistrationDetails = nil
	case "ValidateUserRegistration":
		testFunctions.ValidateUserRegistration = nil
	case "CancelUserRegistration":
		testFunctions.CancelUserRegistration = nil
	}
}

func TestValidateOrCancelRegistration_Validate(t *testing.T) {
	runCases(t, ValidateRegistration, []testCase{
		{name: "GetUserDetailsByID not defined", id: "u", key: "k", setup: registrationValidationStubs{missing: "GetUserDetailsByID"}.install, want: http.StatusBadRequest},
		{name: "GetRegistrationDetails not defined", id: "u", key: "k", setup: registrationValidationStubs{missing: "GetRegistrationDetails"}.install, want: http.StatusBadRequest},
		{name: "ValidateUserRegistration not defined", id: "u", key: "k", setup: registrationValidationStubs{missing: "ValidateUserRegistration"}.install, want: http.StatusBadRequest},
		{name: "CancelUserRegistration not defined", id: "u", key: "k", setup: registrationValidationStubs{missing: "CancelUserRegistration"}.install, want: http.StatusBadRequest},
		{name: "missing key", id: "u", setup: registrationValidationStubs{}.install, want: http.StatusBadRequest},
		{name: "user error", id: "u", key: "k", setup: registrationValidationStubs{userErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "details error", id: "u", key: "k", setup: registrationValidationStubs{detailsErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "validation error", id: "u", key: "k", setup: registrationValidationStubs{validateErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "token creation failure", id: "u", key: "k", setup: registrationValidationStubs{}.install, want: http.StatusBadRequest},
		{name: "success", id: "u", key: "k", setup: all(registrationValidationStubs{}.install, loginStubs), want: http.StatusOK, check: contains("auth_token")},
	})
}

func TestValidateOrCancelRegistration_Cancel(t *testing.T) {
	runCases(t, CancelRegistration, []testCase{
		{name: "missing key", id: "u", setup: registrationValidationStubs{}.install, want: http.StatusBadRequest},
		{name: "user error", id: "u", key: "k", setup: registrationValidationStubs{userErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "cancellation error", id: "u", key: "k", setup: registrationValidationStubs{cancelErr: errBoom}.install, want: http.StatusBadRequest},
		{name: "success", id: "u", key: "k", setup: registrationValidationStubs{}.install, want: http.StatusOK},
	})
}
