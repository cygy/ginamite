package queue

// GetIPLocation : the location of an IP address must be retrieved.
func GetIPLocation(payload TaskIPLocation) {
	CreateTask(MessageTaskIPLocation, payload)
}

// DeleteUser : deletes the data of a user.
func DeleteUser(payload TaskDeleteUser) {
	CreateTask(MessageTaskDeleteUser, payload)
}

// DisableUser : disables the data of a user.
func DisableUser(payload TaskDisableUser) {
	CreateTask(MessageTaskDisableUser, payload)
}

// EnableUser : enables the data of a user.
func EnableUser(payload TaskEnableUser) {
	CreateTask(MessageTaskEnableUser, payload)
}

// UpdateUserSocialNetworks : updates the data of a user.
func UpdateUserSocialNetworks(payload TaskUpdateUserSocialNetworks) {
	CreateTask(MessageTaskUpdateUserSocialNetworks, payload)
}

// RegistrationDone : informs that a new user is registered.
func RegistrationDone(payload TaskRegistrationDone) {
	CreateTask(MessageTaskRegistrationDone, payload)
}

// RegistrationValidated : informs that a new user is validated.
func RegistrationValidated(payload TaskRegistrationValidated) {
	CreateTask(MessageTaskRegistrationValidated, payload)
}
