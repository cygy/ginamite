// Package rights provides the middleware checking the rights of a user.
package rights

import (
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/response"

	"github.com/gin-gonic/gin"
)

// Check checks the rights of a route.
func Check(adminRequired bool, abilities []string, getUserAbilites GetUserAbilitiesFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Do nothing if there is no specific right.
		if len(abilities) == 0 && !adminRequired {
			return
		}

		// Get the user's abilities.
		userID := ctxUser.GetID(c)
		isUserAdmin, userAbilities := getUserAbilites(c, userID)

		locale := ctxLocale.Get(c)

		// Check if the user is an admin.
		if adminRequired && !isUserAdmin {
			response.UnauthorizedWithInsufficientRights(c, locale)
			return
		}

		// Check the abilities of the user.
		for _, requiredAbility := range abilities {
			if !hasUserAbility(requiredAbility, userAbilities) {
				response.UnauthorizedWithInsufficientRights(c, locale)
				return
			}
		}
	}
}
