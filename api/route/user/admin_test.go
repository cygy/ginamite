package user

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/cygy/ginamite/common"
	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/notifications"

	"github.com/gin-gonic/gin"
)

var errBoom = errors.New("boom")

// handlerCase is a scenario of a handler: the JSON body, the injected functions and the expected HTTP status.
type handlerCase struct {
	name  string
	body  string
	setup func()
	want  int
}

func runHandlerCases(t *testing.T, handler gin.HandlerFunc, tests []handlerCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(t)
			if tt.setup != nil {
				tt.setup()
			}

			if got := serve(handler, tt.body); got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetAdminEnabledUsers(t *testing.T) {
	runHandlerCases(t, GetAdminEnabledUsers, []handlerCase{
		{"storage not defined", "", withoutStorage, http.StatusInternalServerError},
		{"error", "", func() {
			testStorage.GetEnabledUsersForAdminFn = func(context.Context) (any, error) { return nil, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testStorage.GetEnabledUsersForAdminFn = func(context.Context) (any, error) { return []string{"u"}, nil }
		}, http.StatusOK},
	})
}

func TestGetAdminDisabledUsers(t *testing.T) {
	runHandlerCases(t, GetAdminDisabledUsers, []handlerCase{
		{"storage not defined", "", withoutStorage, http.StatusInternalServerError},
		{"error", "", func() {
			testStorage.GetDisabledUsersForAdminFn = func(context.Context) (any, error) { return nil, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testStorage.GetDisabledUsersForAdminFn = func(context.Context) (any, error) { return []string{"u"}, nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdatePassword(t *testing.T) {
	var gotUser, gotPassword string
	runHandlerCases(t, AdminUpdatePassword, []handlerCase{
		{"missing password", `{}`, nil, http.StatusBadRequest},
		{"storage not defined", `{"password":"secret"}`, withoutStorage, http.StatusInternalServerError},
		{"error", `{"password":"secret"}`, func() {
			testStorage.UpdatePasswordByAdminFn = func(context.Context, string, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", `{"password":"secret"}`, func() {
			testStorage.UpdatePasswordByAdminFn = func(_ context.Context, userID, password string) error {
				gotUser, gotPassword = userID, password
				return nil
			}
		}, http.StatusOK},
	})

	if gotUser != "42" || gotPassword != "secret" {
		t.Errorf("function called with %q, %q, want 42, secret", gotUser, gotPassword)
	}
}

func TestAdminUpdateInformations(t *testing.T) {
	body := `{"first_name":"John","last_name":"Doe","email_address":"john@example.com","valid_email_address":true}`
	runHandlerCases(t, AdminUpdateInformations, []handlerCase{
		{"missing email address", `{}`, nil, http.StatusBadRequest},
		{"invalid email address", `{"email_address":"john"}`, nil, http.StatusBadRequest},
		{"storage not defined", body, withoutStorage, http.StatusInternalServerError},
		{"error", body, func() {
			testStorage.UpdateInformationsByAdminFn = func(context.Context, string, string, string, string, bool) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testStorage.UpdateInformationsByAdminFn = func(context.Context, string, string, string, string, bool) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdateDescription(t *testing.T) {
	body := `{"description":"hello","locale":"en"}`
	runHandlerCases(t, AdminUpdateDescription, []handlerCase{
		{"missing locale", `{"description":"hello"}`, nil, http.StatusBadRequest},
		{"unsupported locale", `{"locale":"xx"}`, nil, http.StatusBadRequest},
		{"storage not defined", body, withoutStorage, http.StatusInternalServerError},
		{"error", body, func() {
			testStorage.UpdateDescriptionByAdminFn = func(context.Context, string, string, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testStorage.UpdateDescriptionByAdminFn = func(context.Context, string, string, string) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdateSocial(t *testing.T) {
	body := `{"facebook":"` + common.FacebookPrefix + `me","twitter":"` + common.TwitterPrefix + `me","instagram":"` + common.InstagramPrefix + `me"}`
	runHandlerCases(t, AdminUpdateSocial, []handlerCase{
		{"invalid facebook", `{"facebook":"me"}`, nil, http.StatusBadRequest},
		{"invalid twitter", `{"twitter":"me"}`, nil, http.StatusBadRequest},
		{"invalid instagram", `{"instagram":"me"}`, nil, http.StatusBadRequest},
		{"storage not defined", body, withoutStorage, http.StatusInternalServerError},
		{"error", body, func() {
			testStorage.UpdateSocialByAdminFn = func(context.Context, string, string, string, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testStorage.UpdateSocialByAdminFn = func(context.Context, string, string, string, string) error { return nil }
		}, http.StatusOK},
		{"success without network", `{}`, func() {
			testStorage.UpdateSocialByAdminFn = func(context.Context, string, string, string, string) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdateSettings(t *testing.T) {
	body := `{"locale":"en","timezone":"Europe/Paris"}`
	runHandlerCases(t, AdminUpdateSettings, []handlerCase{
		{"missing locale", `{"timezone":"Europe/Paris"}`, nil, http.StatusBadRequest},
		{"missing timezone", `{"locale":"en"}`, nil, http.StatusBadRequest},
		{"unsupported locale", `{"locale":"xx","timezone":"Europe/Paris"}`, nil, http.StatusBadRequest},
		{"unsupported timezone", `{"locale":"en","timezone":"Mars/Base"}`, nil, http.StatusBadRequest},
		{"storage not defined", body, withoutStorage, http.StatusInternalServerError},
		{"error", body, func() {
			testStorage.UpdateSettingsByAdminFn = func(context.Context, string, string, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testStorage.UpdateSettingsByAdminFn = func(context.Context, string, string, string) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdateNotifications(t *testing.T) {
	body := `{"notifications":["` + notifications.TypeNewLogin + `"]}`
	runHandlerCases(t, AdminUpdateNotifications, []handlerCase{
		{"unsupported notification", `{"notifications":["unknown"]}`, nil, http.StatusBadRequest},
		{"storage not defined", body, withoutStorage, http.StatusInternalServerError},
		{"error", body, func() {
			testStorage.UpdateNotificationsByAdminFn = func(context.Context, string, []string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testStorage.UpdateNotificationsByAdminFn = func(context.Context, string, []string) error { return nil }
		}, http.StatusOK},
	})
}

func TestGetAdminUserTokens(t *testing.T) {
	runHandlerCases(t, GetAdminUserTokens, []handlerCase{
		{"storage not defined", "", withoutStorage, http.StatusInternalServerError},
		{"error", "", func() {
			testStorage.GetTokensForAdminFn = func(context.Context, string) (any, error) { return nil, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testStorage.GetTokensForAdminFn = func(context.Context, string) (any, error) { return []string{"t"}, nil }
		}, http.StatusOK},
	})
}

func TestAdminDeleteUserToken(t *testing.T) {
	token := authentication.Token{UserID: "user", ExpirationDate: time.Now().Add(time.Hour)}
	getToken := func(tok authentication.Token, err error) func() {
		return func() {
			testTokens.detailsFn = func(context.Context, string) (authentication.Token, error) { return tok, err }
		}
	}
	deleteToken := func(err error) func() {
		return func() {
			testStorage.DeleteTokenByIDByAdminFn = func(context.Context, string) error { return err }
		}
	}
	both := func(a, b func()) func() { return func() { a(); b() } }

	runHandlerCases(t, AdminDeleteUserToken, []handlerCase{
		{"storage not defined", "", both(withoutStorage, getToken(token, nil)), http.StatusInternalServerError},
		{"tokens storage not defined", "", both(deleteToken(nil), func() { testAccountFunctions.Tokens = nil }), http.StatusInternalServerError},
		{"details error", "", both(deleteToken(nil), getToken(token, errBoom)), http.StatusInternalServerError},
		{"token not found", "", both(deleteToken(nil), getToken(authentication.Token{}, commonErrors.ErrNotFound)), http.StatusOK},
		{"delete error", "", both(deleteToken(errBoom), getToken(token, nil)), http.StatusInternalServerError},
		{"delete not found", "", both(deleteToken(commonErrors.ErrNotFound), getToken(token, nil)), http.StatusOK},
		{"success", "", both(deleteToken(nil), getToken(token, nil)), http.StatusOK},
	})
}

func TestAdminDelete(t *testing.T) {
	user := authentication.User{ID: "42", EmailAddress: "john@example.com", Username: "john", Locale: "en"}
	runHandlerCases(t, AdminDelete, []handlerCase{
		{"details error", "", func() {
			testAccounts.detailsFn = func(context.Context, string) (authentication.User, error) { return user, errBoom }
		}, http.StatusInternalServerError},
		{"delete error", "", func() {
			testAccounts.detailsFn = func(context.Context, string) (authentication.User, error) { return user, nil }
			testAccounts.deleteFn = func(context.Context, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testAccounts.detailsFn = func(context.Context, string) (authentication.User, error) { return user, nil }
			testAccounts.deleteFn = func(context.Context, string) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminDisable(t *testing.T) {
	user := authentication.User{ID: "42", EmailAddress: "john@example.com", Username: "john", Locale: "en"}
	runHandlerCases(t, AdminDisable, []handlerCase{
		{"error", "", func() {
			testAccounts.disableFn = func(context.Context, string) (authentication.User, error) { return user, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testAccounts.disableFn = func(context.Context, string) (authentication.User, error) { return user, nil }
		}, http.StatusOK},
	})
}

func TestAdminEnable(t *testing.T) {
	user := authentication.User{ID: "42", EmailAddress: "john@example.com", Username: "john", Locale: "en"}
	runHandlerCases(t, AdminEnable, []handlerCase{
		{"error", "", func() {
			testAccounts.enableFn = func(context.Context, string) (authentication.User, error) { return user, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testAccounts.enableFn = func(context.Context, string) (authentication.User, error) { return user, nil }
		}, http.StatusOK},
	})
}
