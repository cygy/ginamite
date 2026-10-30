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
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	// Here is the function to delete the user and which return the auth tokens revoked.
	// Each token is defined with its unique ID and its expiration date.
	revokedAuthTokens := deleteUserByIDFunc(context.TODO(), p.UserID)

	if len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}
}

// DisableByID : disables a user with a ID.
func DisableByID(payload []byte, disableUserByIDFunc func(ctx context.Context, userID string) map[string]time.Time) {
	if disableUserByIDFunc == nil {
		return
	}

	p := queue.TaskDisableUser{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	// Here is the function to disable the user and which return the auth tokens revoked.
	// Each token is defined with its unique ID and its expiration date.
	revokedAuthTokens := disableUserByIDFunc(context.TODO(), p.UserID)

	if len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}
}

// EnableByID : enables a user with a ID.
func EnableByID(payload []byte, enableUserByIDFunc func(ctx context.Context, userID string)) {
	if enableUserByIDFunc == nil {
		return
	}

	p := queue.TaskDisableUser{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	enableUserByIDFunc(context.TODO(), p.UserID)
}

// UpdateSocialNetworksByID : updates the documents referencing the social networks of a user.
func UpdateSocialNetworksByID(payload []byte, updateUserSocialNetworksByIDFunc func(ctx context.Context, userID string)) {
	if updateUserSocialNetworksByIDFunc == nil {
		return
	}

	p := queue.TaskUpdateUserSocialNetworks{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	// Here is the function to update the documents referencing the social networks of a user.
	updateUserSocialNetworksByIDFunc(context.TODO(), p.UserID)
}

// RegistrationDone : informs that a user is registered.
func RegistrationDone(payload []byte, registrationDoneFunc func(ctx context.Context, userID string)) {
	if registrationDoneFunc == nil {
		return
	}

	p := queue.TaskRegistrationDone{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	// Here is the function to inform that a user is registered.
	registrationDoneFunc(context.TODO(), p.UserID)
}

// RegistrationValidated : informs that a user is validated.
func RegistrationValidated(payload []byte, registrationValidatedFunc func(ctx context.Context, userID string)) {
	if registrationValidatedFunc == nil {
		return
	}

	p := queue.TaskRegistrationValidated{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	// Here is the function to inform that a user is registered.
	registrationValidatedFunc(context.TODO(), p.UserID)
}
