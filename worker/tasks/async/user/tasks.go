package user

import (
	"time"

	"github.com/cygy/ginamite/common/queue"
)

// DeleteByID : deletes a user with a ID.
func DeleteByID(payload []byte, deleteUserByIDFunc func(userID string) map[string]time.Time) {
	if deleteUserByIDFunc == nil {
		return
	}

	p := queue.TaskDeleteUser{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to delete the user and which return the auth tokens revoked.
	// Each token is defined with its unique ID and its expiration date.
	revokedAuthTokens := deleteUserByIDFunc(p.UserID)

	if revokedAuthTokens != nil && len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}
}

// DisableByID : disables a user with a ID.
func DisableByID(payload []byte, disableUserByIDFunc func(userID string) map[string]time.Time) {
	if disableUserByIDFunc == nil {
		return
	}

	p := queue.TaskDisableUser{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to disable the user and which return the auth tokens revoked.
	// Each token is defined with its unique ID and its expiration date.
	revokedAuthTokens := disableUserByIDFunc(p.UserID)

	if revokedAuthTokens != nil && len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}
}

// EnableByID : enables a user with a ID.
func EnableByID(payload []byte, enableUserByIDFunc func(userID string)) {
	if enableUserByIDFunc == nil {
		return
	}

	p := queue.TaskDisableUser{}
	queue.UnmarshalPayload(payload, &p)

	enableUserByIDFunc(p.UserID)
}

// UpdateSocialNetworksByID : updates the documents referencing the social networks of a user.
func UpdateSocialNetworksByID(payload []byte, updateUserSocialNetworksByIDFunc func(userID string)) {
	if updateUserSocialNetworksByIDFunc == nil {
		return
	}

	p := queue.TaskUpdateUserSocialNetworks{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to update the documents referencing the social networks of a user.
	updateUserSocialNetworksByIDFunc(p.UserID)
}

// RegistrationDone : informs that a user is registered.
func RegistrationDone(payload []byte, registrationDoneFunc func(userID string)) {
	if registrationDoneFunc == nil {
		return
	}

	p := queue.TaskRegistrationDone{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to inform that a user is registered.
	registrationDoneFunc(p.UserID)
}

// RegistrationValidated : informs that a user is validated.
func RegistrationValidated(payload []byte, registrationValidatedFunc func(userID string)) {
	if registrationValidatedFunc == nil {
		return
	}

	p := queue.TaskRegistrationValidated{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to inform that a user is registered.
	registrationValidatedFunc(p.UserID)
}
