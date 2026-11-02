package response

import (
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/common/response"
)

func TestAddRenewedAuthToken(t *testing.T) {
	c, recorder := apitest.NewContext("GET", "/", nil)
	AddRenewedAuthToken(c, "new-token")

	if got := recorder.Header().Get(response.RenewedAuthTokenHeader); got != "new-token" {
		t.Errorf("header = %q, want new-token", got)
	}
}
