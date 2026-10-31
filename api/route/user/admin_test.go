package user

import (
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
		{"function not defined", "", nil, http.StatusInternalServerError},
		{"error", "", func() {
			testFunctions.GetEnabledUsersForAdmin = func(*gin.Context) (any, error) { return nil, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testFunctions.GetEnabledUsersForAdmin = func(*gin.Context) (any, error) { return []string{"u"}, nil }
		}, http.StatusOK},
	})
}

func TestGetAdminDisabledUsers(t *testing.T) {
	runHandlerCases(t, GetAdminDisabledUsers, []handlerCase{
		{"function not defined", "", nil, http.StatusInternalServerError},
		{"error", "", func() {
			testFunctions.GetDisabledUsersForAdmin = func(*gin.Context) (any, error) { return nil, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testFunctions.GetDisabledUsersForAdmin = func(*gin.Context) (any, error) { return []string{"u"}, nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdatePassword(t *testing.T) {
	var gotUser, gotPassword string
	runHandlerCases(t, AdminUpdatePassword, []handlerCase{
		{"missing password", `{}`, nil, http.StatusBadRequest},
		{"function not defined", `{"password":"secret"}`, nil, http.StatusInternalServerError},
		{"error", `{"password":"secret"}`, func() {
			testFunctions.UpdatePasswordByAdmin = func(*gin.Context, string, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", `{"password":"secret"}`, func() {
			testFunctions.UpdatePasswordByAdmin = func(_ *gin.Context, userID, password string) error {
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
		{"function not defined", body, nil, http.StatusInternalServerError},
		{"error", body, func() {
			testFunctions.UpdateInformationsByAdmin = func(*gin.Context, string, string, string, string, bool) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testFunctions.UpdateInformationsByAdmin = func(*gin.Context, string, string, string, string, bool) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdateDescription(t *testing.T) {
	body := `{"description":"hello","locale":"en"}`
	runHandlerCases(t, AdminUpdateDescription, []handlerCase{
		{"missing locale", `{"description":"hello"}`, nil, http.StatusBadRequest},
		{"unsupported locale", `{"locale":"xx"}`, nil, http.StatusBadRequest},
		{"function not defined", body, nil, http.StatusInternalServerError},
		{"error", body, func() {
			testFunctions.UpdateDescriptionByAdmin = func(*gin.Context, string, string, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testFunctions.UpdateDescriptionByAdmin = func(*gin.Context, string, string, string) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdateSocial(t *testing.T) {
	body := `{"facebook":"` + common.FacebookPrefix + `me","twitter":"` + common.TwitterPrefix + `me","instagram":"` + common.InstagramPrefix + `me"}`
	runHandlerCases(t, AdminUpdateSocial, []handlerCase{
		{"invalid facebook", `{"facebook":"me"}`, nil, http.StatusBadRequest},
		{"invalid twitter", `{"twitter":"me"}`, nil, http.StatusBadRequest},
		{"invalid instagram", `{"instagram":"me"}`, nil, http.StatusBadRequest},
		{"function not defined", body, nil, http.StatusInternalServerError},
		{"error", body, func() {
			testFunctions.UpdateSocialByAdmin = func(*gin.Context, string, string, string, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testFunctions.UpdateSocialByAdmin = func(*gin.Context, string, string, string, string) error { return nil }
		}, http.StatusOK},
		{"success without network", `{}`, func() {
			testFunctions.UpdateSocialByAdmin = func(*gin.Context, string, string, string, string) error { return nil }
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
		{"function not defined", body, nil, http.StatusInternalServerError},
		{"error", body, func() {
			testFunctions.UpdateSettingsByAdmin = func(*gin.Context, string, string, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testFunctions.UpdateSettingsByAdmin = func(*gin.Context, string, string, string) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminUpdateNotifications(t *testing.T) {
	body := `{"notifications":["` + notifications.TypeNewLogin + `"]}`
	runHandlerCases(t, AdminUpdateNotifications, []handlerCase{
		{"unsupported notification", `{"notifications":["unknown"]}`, nil, http.StatusBadRequest},
		{"function not defined", body, nil, http.StatusInternalServerError},
		{"error", body, func() {
			testFunctions.UpdateNotificationsByAdmin = func(*gin.Context, string, []string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", body, func() {
			testFunctions.UpdateNotificationsByAdmin = func(*gin.Context, string, []string) error { return nil }
		}, http.StatusOK},
	})
}

func TestGetAdminUserTokens(t *testing.T) {
	runHandlerCases(t, GetAdminUserTokens, []handlerCase{
		{"function not defined", "", nil, http.StatusInternalServerError},
		{"error", "", func() {
			testFunctions.GetTokensForAdmin = func(*gin.Context, string) (any, error) { return nil, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testFunctions.GetTokensForAdmin = func(*gin.Context, string) (any, error) { return []string{"t"}, nil }
		}, http.StatusOK},
	})
}

func TestAdminDeleteUserToken(t *testing.T) {
	token := authentication.Token{UserID: "user", ExpirationDate: time.Now().Add(time.Hour)}
	getToken := func(tok authentication.Token, err error) func() {
		return func() {
			testAccountFunctions.GetTokenDetailsByID = func(*gin.Context, string) (authentication.Token, error) { return tok, err }
		}
	}
	deleteToken := func(err error) func() {
		return func() {
			testFunctions.DeleteTokenByIDByAdmin = func(*gin.Context, string) error { return err }
		}
	}
	both := func(a, b func()) func() { return func() { a(); b() } }

	runHandlerCases(t, AdminDeleteUserToken, []handlerCase{
		{"delete function not defined", "", getToken(token, nil), http.StatusInternalServerError},
		{"details function not defined", "", deleteToken(nil), http.StatusInternalServerError},
		{"details error", "", both(deleteToken(nil), getToken(token, errBoom)), http.StatusInternalServerError},
		{"token not found", "", both(deleteToken(nil), getToken(authentication.Token{}, commonErrors.NotFound())), http.StatusOK},
		{"delete error", "", both(deleteToken(errBoom), getToken(token, nil)), http.StatusInternalServerError},
		{"delete not found", "", both(deleteToken(commonErrors.NotFound()), getToken(token, nil)), http.StatusOK},
		{"success", "", both(deleteToken(nil), getToken(token, nil)), http.StatusOK},
	})
}

func TestAdminDelete(t *testing.T) {
	user := authentication.User{ID: "42", EmailAddress: "john@example.com", Username: "john", Locale: "en"}
	runHandlerCases(t, AdminDelete, []handlerCase{
		{"details error", "", func() {
			testAccountFunctions.GetUserDetailsByID = func(*gin.Context, string) (authentication.User, error) { return user, errBoom }
		}, http.StatusInternalServerError},
		{"delete error", "", func() {
			testAccountFunctions.GetUserDetailsByID = func(*gin.Context, string) (authentication.User, error) { return user, nil }
			testAccountFunctions.DeleteUserByID = func(*gin.Context, string) error { return errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testAccountFunctions.GetUserDetailsByID = func(*gin.Context, string) (authentication.User, error) { return user, nil }
			testAccountFunctions.DeleteUserByID = func(*gin.Context, string) error { return nil }
		}, http.StatusOK},
	})
}

func TestAdminDisable(t *testing.T) {
	user := authentication.User{ID: "42", EmailAddress: "john@example.com", Username: "john", Locale: "en"}
	runHandlerCases(t, AdminDisable, []handlerCase{
		{"error", "", func() {
			testAccountFunctions.DisableUserByID = func(*gin.Context, string) (authentication.User, error) { return user, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testAccountFunctions.DisableUserByID = func(*gin.Context, string) (authentication.User, error) { return user, nil }
		}, http.StatusOK},
	})
}

func TestAdminEnable(t *testing.T) {
	user := authentication.User{ID: "42", EmailAddress: "john@example.com", Username: "john", Locale: "en"}
	runHandlerCases(t, AdminEnable, []handlerCase{
		{"error", "", func() {
			testAccountFunctions.EnableUserByID = func(*gin.Context, string) (authentication.User, error) { return user, errBoom }
		}, http.StatusInternalServerError},
		{"success", "", func() {
			testAccountFunctions.EnableUserByID = func(*gin.Context, string) (authentication.User, error) { return user, nil }
		}, http.StatusOK},
	})
}
