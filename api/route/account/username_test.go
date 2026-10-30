package account

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetAvailability(t *testing.T) {
	exists := func(exists bool, err error) func() {
		return func() {
			DoesAccountWithUsernameExist = func(*gin.Context, string) (bool, error) { return exists, err }
		}
	}

	runCases(t, GetAvailability, []testCase{
		{name: "missing username", body: `{}`, want: http.StatusBadRequest},
		{name: "invalid username", body: `{"username":"ab"}`, want: http.StatusBadRequest},
		{name: "no lookup function", body: `{"username":"Cyril"}`, want: http.StatusOK, check: contains(`"available":true`)},
		{name: "free username", body: `{"username":"Cyril"}`, setup: exists(false, nil), want: http.StatusOK, check: contains(`"available":true`)},
		{name: "taken username", body: `{"username":"Cyril"}`, setup: exists(true, nil), want: http.StatusOK, check: contains(`"available":false`)},
		{name: "lookup error", body: `{"username":"Cyril"}`, setup: exists(false, errBoom), want: http.StatusOK, check: contains(`"available":false`)},
	})
}
