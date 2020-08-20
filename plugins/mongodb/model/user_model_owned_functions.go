package model

// HasAbility : returns true if the user has the ability.
func (user *OwnUser) HasAbility(ability string) bool {
	for _, a := range user.Abilities {
		if a == ability {
			return true
		}
	}

	return false
}
