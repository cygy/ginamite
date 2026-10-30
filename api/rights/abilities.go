package rights

import "slices"

func hasUserAbility(requiredAbility string, userAbilities []string) bool {
	return slices.Contains(userAbilities, requiredAbility)
}
