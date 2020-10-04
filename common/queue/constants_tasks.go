package queue

const (
	// MessageTaskIPLocation : gets the location of an IP address.
	MessageTaskIPLocation = "ip_location"

	// MessageTaskDeleteUser : deletes the data of a user.
	MessageTaskDeleteUser = "delete_user"

	// MessageTaskDisableUser : disables the data of a user.
	MessageTaskDisableUser = "disable_user"

	// MessageTaskEnableUser : enables the data of a user.
	MessageTaskEnableUser = "enable_user"

	// MessageTaskUpdateUserSocialNetworks : updates the social networks of a user.
	MessageTaskUpdateUserSocialNetworks = "update_user_social_networks"

	// MessageTaskRegistrationDone : informs that a new user is registered.
	MessageTaskRegistrationDone = "registration_done"

	// MessageTaskRegistrationValidated : informs that a new user is validated.
	MessageTaskRegistrationValidated = "registration_validated"
)
