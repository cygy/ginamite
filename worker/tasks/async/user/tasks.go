package user

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/queue"
)

// DeleteByID : deletes a user with a ID.
func DeleteByID(payload []byte, deleteUserByIDFunc func(ctx context.Context, userID string) map[string]time.Time) {
	if deleteUserByIDFunc == nil {
		return
	}

	p := queue.TaskDeleteUser{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to delete the user and which return the auth tokens revoked.
	// Each token is defined with its unique ID and its expiration date.
	revokedAuthTokens := deleteUserByIDFunc(nil, p.UserID)

	if revokedAuthTokens != nil && len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}
}

// DisableByID : disables a user with a ID.
func DisableByID(payload []byte, disableUserByIDFunc func(ctx context.Context, userID string) map[string]time.Time) {
	if disableUserByIDFunc == nil {
		return
	}

	p := queue.TaskDisableUser{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to disable the user and which return the auth tokens revoked.
	// Each token is defined with its unique ID and its expiration date.
	revokedAuthTokens := disableUserByIDFunc(nil, p.UserID)

	if revokedAuthTokens != nil && len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}
}

// EnableByID : enables a user with a ID.
func EnableByID(payload []byte, enableUserByIDFunc func(ctx context.Context, userID string)) {
	if enableUserByIDFunc == nil {
		return
	}

	p := queue.TaskDisableUser{}
	queue.UnmarshalPayload(payload, &p)

	enableUserByIDFunc(nil, p.UserID)
}

// UpdateSocialNetworksByID : updates the documents referencing the social networks of a user.
func UpdateSocialNetworksByID(payload []byte, updateUserSocialNetworksByIDFunc func(ctx context.Context, userID string)) {
	if updateUserSocialNetworksByIDFunc == nil {
		return
	}

	p := queue.TaskUpdateUserSocialNetworks{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to update the documents referencing the social networks of a user.
	updateUserSocialNetworksByIDFunc(nil, p.UserID)
}

// RegistrationDone : informs that a user is registered.
func RegistrationDone(payload []byte, registrationDoneFunc func(ctx context.Context, userID string)) {
	if registrationDoneFunc == nil {
		return
	}

	p := queue.TaskRegistrationDone{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to inform that a user is registered.
	registrationDoneFunc(nil, p.UserID)
}

// RegistrationValidated : informs that a user is validated.
func RegistrationValidated(payload []byte, registrationValidatedFunc func(ctx context.Context, userID string)) {
	if registrationValidatedFunc == nil {
		return
	}

	p := queue.TaskRegistrationValidated{}
	queue.UnmarshalPayload(payload, &p)

	// Here is the function to inform that a user is registered.
	registrationValidatedFunc(nil, p.UserID)
}
