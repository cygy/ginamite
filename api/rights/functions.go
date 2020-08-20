package rights

import (
	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"

	"github.com/gin-gonic/gin"
)

// CheckRights : checks the rights of a route.
func CheckRights(r Rights, getUserAbilites GetUserAbilitesFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Do nothing if there is no specific right.
		if len(r.Abilities) == 0 && !r.AdminRequired {
			return
		}

		// Get the user's abilities.
		userID := context.GetUserID(c)
		isUserAdmin, userAbilities := getUserAbilites(c, userID)

		locale := context.GetLocale(c)

		// Check if the user is an admin.
		if r.AdminRequired && !isUserAdmin {
			response.UnauthorizedWithInsufficientRights(c, locale)
			return
		}

		// Check the abilities of the user.
		requiredAbilities := r.Abilities
		for _, requiredAbility := range requiredAbilities {
			if !hasUserAbility(requiredAbility, userAbilities) {
				response.UnauthorizedWithInsufficientRights(c, locale)
				return
			}
		}
	}
}

// ----- Helpers
func hasUserAbility(requiredAbility string, userAbilities []string) bool {
	for _, userAbility := range userAbilities {
		if userAbility == requiredAbility {
			return true
		}
	}

	return false
}
