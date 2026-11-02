package account

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"

	"github.com/gin-gonic/gin"
)

func TestLoginWithThirdParty(t *testing.T) {
	const body = `{"token":"third-party-token",` + loginDevice + `}`
	infos := validator.TokenInfos{UserID: "social-id", Email: "john@example.com"}

	for _, method := range []string{authentication.MethodFacebook, authentication.MethodGoogle} {
		valid := withValidator(fakeValidator{source: method, infos: infos, ok: true})
		bySocialID := func(user authentication.User, err error) func() {
			return func() {
				fn := func(*gin.Context, string) (authentication.User, error) { return user, err }
				if method == authentication.MethodFacebook {
					testFunctions.GetUserDetailsByFacebookUserID = fn
				} else {
					testFunctions.GetUserDetailsByGoogleUserID = fn
				}
			}
		}
		byIdentifier := func(user authentication.User, err error) func() {
			return func() {
				testFunctions.GetUserDetailsByIdentifier = func(*gin.Context, string) (authentication.User, error) { return user, err }
			}
		}
		save := func(err error) func() {
			return func() {
				fn := func(*gin.Context, string, validator.TokenInfos) error { return err }
				if method == authentication.MethodFacebook {
					testFunctions.SaveFacebookUserTokenDetailsToUser = fn
				} else {
					testFunctions.SaveGoogleUserTokenDetailsToUser = fn
				}
			}
		}
		notFound := commonErrors.NotFound()
		handler := func(c *gin.Context) { loginWithThirdParty(c, newTokenValidator(c, method)) }

		t.Run(method, func(t *testing.T) {
			runCases(t, handler, []testCase{
				{name: "missing token", body: `{` + loginDevice + `}`, setup: valid, want: http.StatusBadRequest},
				{name: "missing source", body: `{"token":"x","device":{"type":"mobile","name":"iPhone"}}`, setup: valid, want: http.StatusBadRequest},
				{name: "missing device", body: `{"token":"x","source":"web"}`, setup: valid, want: http.StatusBadRequest},
				{name: "invalid third-party token", body: body, setup: withValidator(fakeValidator{source: method}), want: http.StatusBadRequest},
				{name: "lookup function not defined", body: body, setup: valid, want: http.StatusBadRequest},
				{name: "lookup error", body: body, setup: all(valid, bySocialID(authentication.User{}, errBoom)), want: http.StatusBadRequest},
				{name: "identifier function not defined", body: body, setup: all(valid, bySocialID(authentication.User{}, notFound)), want: http.StatusBadRequest},
				{name: "identifier error", body: body, setup: all(valid, bySocialID(authentication.User{}, notFound), byIdentifier(authentication.User{}, errBoom)), want: http.StatusBadRequest},
				{name: "unknown user", body: body, setup: all(valid, bySocialID(authentication.User{}, notFound), byIdentifier(authentication.User{}, notFound)), want: http.StatusNotFound},
				{name: "save function not defined", body: body, setup: all(valid, bySocialID(testUser, nil)), want: http.StatusBadRequest},
				{name: "save error", body: body, setup: all(valid, bySocialID(testUser, nil), save(errBoom)), want: http.StatusBadRequest},
				{name: "token creation failure", body: body, setup: all(valid, bySocialID(testUser, nil), save(nil), withoutTokens), want: http.StatusBadRequest},
				{name: "success with the social network ID", body: body, setup: all(valid, bySocialID(testUser, nil), save(nil), loginStubs), want: http.StatusCreated, check: contains("auth_token")},
				{name: "success with the email address", body: body, setup: all(valid, bySocialID(authentication.User{}, notFound), byIdentifier(testUser, nil), save(nil), loginStubs), want: http.StatusCreated},
			})
		})
	}
}

func TestLoginWithFacebookAndGoogle(t *testing.T) {
	for name, handler := range map[string]gin.HandlerFunc{"facebook": LoginWithFacebook, "google": LoginWithGoogle} {
		t.Run(name, func(t *testing.T) {
			runCases(t, handler, []testCase{
				{name: "missing token", body: `{}`, want: http.StatusBadRequest},
				{name: "invalid third-party token", body: `{"token":"x",` + loginDevice + `}`, setup: withValidator(fakeValidator{}), want: http.StatusBadRequest},
			})
		})
	}
}

func TestCanNotLoginError(t *testing.T) {
	setup(t)

	c, recorder := newContext(nil)
	canNotLoginError(c, errBoom)
	c.Writer.WriteHeaderNow()

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if len(c.Errors) != 1 {
		t.Errorf("errors = %v, want the reported error", c.Errors)
	}
}
