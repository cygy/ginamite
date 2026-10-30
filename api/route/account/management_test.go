package account

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetInfos(t *testing.T) {
	details := func(v any) func() {
		return func() { GetOwnedAccountDetails = func(*gin.Context, string) any { return v } }
	}

	runCases(t, GetInfos, []testCase{
		{name: "function not defined", userID: testUserID, want: http.StatusInternalServerError},
		{name: "no details", userID: testUserID, setup: details(nil), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, setup: details(map[string]string{"username": "john"}), want: http.StatusOK, check: contains("profile", "john")},
	})
}

func TestUpdateInfos(t *testing.T) {
	const body = `{"first_name":"John","last_name":"Doe"}`
	update := func(err error) func() {
		return func() { UpdatedAccountInfos = func(*gin.Context, string, string, string) error { return err } }
	}

	runCases(t, UpdateInfos, []testCase{
		{name: "function not defined", userID: testUserID, body: body, want: http.StatusInternalServerError},
		{name: "missing first name", userID: testUserID, body: `{"last_name":"Doe"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "missing last name", userID: testUserID, body: `{"first_name":"John"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "update error", userID: testUserID, body: body, setup: update(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: update(nil), want: http.StatusOK},
	})
}

func TestUpdateSettings(t *testing.T) {
	const body = `{"locale":"en","timezone":"Europe/Paris"}`
	update := func(err error) func() {
		return func() { UpdatedAccountSettings = func(*gin.Context, string, string, string) error { return err } }
	}

	runCases(t, UpdateSettings, []testCase{
		{name: "function not defined", userID: testUserID, body: body, want: http.StatusInternalServerError},
		{name: "missing locale", userID: testUserID, body: `{"timezone":"Europe/Paris"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "missing timezone", userID: testUserID, body: `{"locale":"en"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "unsupported locale", userID: testUserID, body: `{"locale":"xx","timezone":"Europe/Paris"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "unsupported timezone", userID: testUserID, body: `{"locale":"en","timezone":"Mars/Base"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "update error", userID: testUserID, body: body, setup: update(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: update(nil), want: http.StatusOK},
	})
}
