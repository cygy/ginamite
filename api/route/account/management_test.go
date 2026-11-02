package account

import (
	"context"
	"net/http"
	"testing"
)

func TestGetInfos(t *testing.T) {
	details := func(v any, err error) func() {
		return func() {
			testAccounts.GetOwnedAccountDetailsFn = func(context.Context, string) (any, error) { return v, err }
		}
	}

	runCases(t, GetInfos, []testCase{
		{name: "storage of the accounts not defined", userID: testUserID, setup: withoutAccounts, want: http.StatusInternalServerError},
		{name: "error", userID: testUserID, setup: details(nil, errBoom), want: http.StatusInternalServerError},
		{name: "no details", userID: testUserID, setup: details(nil, nil), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, setup: details(map[string]string{"username": "john"}, nil), want: http.StatusOK, check: contains("profile", "john")},
	})
}

func TestUpdateInfos(t *testing.T) {
	const body = `{"first_name":"John","last_name":"Doe"}`
	update := func(err error) func() {
		return func() {
			testAccounts.UpdatedAccountInfosFn = func(context.Context, string, string, string) error { return err }
		}
	}

	runCases(t, UpdateInfos, []testCase{
		{name: "storage of the accounts not defined", userID: testUserID, body: body, setup: withoutAccounts, want: http.StatusInternalServerError},
		{name: "missing first name", userID: testUserID, body: `{"last_name":"Doe"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "missing last name", userID: testUserID, body: `{"first_name":"John"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "update error", userID: testUserID, body: body, setup: update(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: update(nil), want: http.StatusOK},
	})
}

func TestUpdateSettings(t *testing.T) {
	const body = `{"locale":"en","timezone":"Europe/Paris"}`
	update := func(err error) func() {
		return func() {
			testAccounts.UpdatedAccountSettingsFn = func(context.Context, string, string, string) error { return err }
		}
	}

	runCases(t, UpdateSettings, []testCase{
		{name: "storage of the accounts not defined", userID: testUserID, body: body, setup: withoutAccounts, want: http.StatusInternalServerError},
		{name: "missing locale", userID: testUserID, body: `{"timezone":"Europe/Paris"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "missing timezone", userID: testUserID, body: `{"locale":"en"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "unsupported locale", userID: testUserID, body: `{"locale":"xx","timezone":"Europe/Paris"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "unsupported timezone", userID: testUserID, body: `{"locale":"en","timezone":"Mars/Base"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "update error", userID: testUserID, body: body, setup: update(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: update(nil), want: http.StatusOK},
	})
}
