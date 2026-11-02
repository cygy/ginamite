package rights

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

func TestCheck(t *testing.T) {
	const secret = "secret"

	apitest.InitLocalization(t, map[string]string{"placeholder": "placeholder"})

	userAbilities := func(isAdmin bool, abilities ...string) GetUserAbilitiesFunc {
		return func(*gin.Context, string) (bool, []string) { return isAdmin, abilities }
	}

	tests := []struct {
		name          string
		adminRequired bool
		abilities     []string
		user          GetUserAbilitiesFunc
		wantStatus    int
	}{
		{"no specific right", false, nil, nil, http.StatusOK},
		{"admin required and user is admin", true, nil, userAbilities(true), http.StatusOK},
		{"admin required and user is not admin", true, nil, userAbilities(false), http.StatusUnauthorized},
		{"ability held", false, []string{AbilityToManageUsers}, userAbilities(false, AbilityToManageUsers), http.StatusOK},
		{"ability missing", false, []string{AbilityToManageUsers}, userAbilities(false, AbilityToManageContacts), http.StatusUnauthorized},
		{"one ability missing among several", false, []string{AbilityToManageUsers, AbilityToManageContacts}, userAbilities(false, AbilityToManageUsers), http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ctxkey.Localization.Inject(apitest.Localization), ctxLocale.Inject("en", []string{"en"}), ctxAuthToken.Parse(secret, nil, false))
			router.Use(Check(tt.adminRequired, tt.abilities, tt.user))
			router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+apitest.SignToken(t, secret, "user-1", "token-1", false))

			if got := apitest.Serve(router, req).Code; got != tt.wantStatus {
				t.Errorf("status = %d, want %d", got, tt.wantStatus)
			}
		})
	}
}
