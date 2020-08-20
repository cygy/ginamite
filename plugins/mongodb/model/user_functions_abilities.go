package model

// AddUserAbilities : add some abilities to the user abilities list.
func AddUserAbilities(abilities []string) {
	AllUserAbilities = append(AllUserAbilities, abilities...)
}

// AddDefaultUserAbilities : add some abilities to the default user abilities list.
func AddDefaultUserAbilities(abilities []string) {
	DefaultUserAbilities = append(DefaultUserAbilities, abilities...)
}
