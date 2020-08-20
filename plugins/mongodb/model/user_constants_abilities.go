package model

import "github.com/cygy/ginamite/api/rights"

var (
	// AllUserAbilities : all the user abilities are listed here.
	AllUserAbilities = []string{
		rights.AbilityToManageUsers,
		rights.AbilityToManageContacts,
	}

	// DefaultUserAbilities : these are the default abilities of the users.
	DefaultUserAbilities = []string{}
)
