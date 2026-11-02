package account

import (
	"context"
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common"
)

func TestUpdateSocialNetworks(t *testing.T) {
	body := `{"facebook":"` + common.FacebookPrefix + `me","twitter":"` + common.TwitterPrefix + `me","instagram":"` + common.InstagramPrefix + `me"}`
	update := func(err error) func() {
		return func() {
			testAccounts.UpdateUserSocialNetworksFn = func(context.Context, string, string, string, string) error { return err }
		}
	}

	runCases(t, UpdateSocialNetworks, []testCase{
		{name: "storage of the accounts not defined", userID: testUserID, body: body, setup: withoutAccounts, want: http.StatusInternalServerError},
		{name: "invalid facebook", userID: testUserID, body: `{"facebook":"me"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "invalid twitter", userID: testUserID, body: `{"twitter":"me"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "invalid instagram", userID: testUserID, body: `{"instagram":"me"}`, setup: update(nil), want: http.StatusBadRequest},
		{name: "update error", userID: testUserID, body: body, setup: update(errBoom), want: http.StatusInternalServerError},
		{name: "success", userID: testUserID, body: body, setup: update(nil), want: http.StatusOK},
		{name: "success without network", userID: testUserID, body: `{}`, setup: update(nil), want: http.StatusOK},
	})
}
