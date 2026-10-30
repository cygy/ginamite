// Package user provides the asynchronous tasks about the users.
package user

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/queue"
)

// DeleteByID decodes the payload of a task to delete a user and calls deleteUserByIDFunc, which deletes the user
// and returns the auth tokens revoked, defined by their ID and their expiration date. The tokens are then invalidated.
// Nothing happens if the function is nil.
func DeleteByID(payload []byte, deleteUserByIDFunc func(ctx context.Context, userID string) map[string]time.Time) {
	if deleteUserByIDFunc == nil {
		return
	}

	p := queue.TaskDeleteUser{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	if revokedAuthTokens := deleteUserByIDFunc(context.Background(), p.UserID); len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}
}

// DisableByID decodes the payload of a task to disable a user and calls disableUserByIDFunc, which disables the user
// and returns the auth tokens revoked, defined by their ID and their expiration date. The tokens are then invalidated.
// Nothing happens if the function is nil.
func DisableByID(payload []byte, disableUserByIDFunc func(ctx context.Context, userID string) map[string]time.Time) {
	if disableUserByIDFunc == nil {
		return
	}

	p := queue.TaskDisableUser{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	if revokedAuthTokens := disableUserByIDFunc(context.Background(), p.UserID); len(revokedAuthTokens) > 0 {
		queue.InvalidAuthTokens(revokedAuthTokens)
	}
}

// EnableByID decodes the payload of a task to enable a user and calls enableUserByIDFunc.
// Nothing happens if the function is nil.
func EnableByID(payload []byte, enableUserByIDFunc func(ctx context.Context, userID string)) {
	if enableUserByIDFunc == nil {
		return
	}

	p := queue.TaskEnableUser{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	enableUserByIDFunc(context.Background(), p.UserID)
}

// UpdateSocialNetworksByID decodes the payload of a task and calls updateUserSocialNetworksByIDFunc, which updates
// the documents referencing the social networks of a user. Nothing happens if the function is nil.
func UpdateSocialNetworksByID(payload []byte, updateUserSocialNetworksByIDFunc func(ctx context.Context, userID string)) {
	if updateUserSocialNetworksByIDFunc == nil {
		return
	}

	p := queue.TaskUpdateUserSocialNetworks{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	updateUserSocialNetworksByIDFunc(context.Background(), p.UserID)
}

// RegistrationDone decodes the payload of a task and calls registrationDoneFunc, which handles a user who registered.
// Nothing happens if the function is nil.
func RegistrationDone(payload []byte, registrationDoneFunc func(ctx context.Context, userID string)) {
	if registrationDoneFunc == nil {
		return
	}

	p := queue.TaskRegistrationDone{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	registrationDoneFunc(context.Background(), p.UserID)
}

// RegistrationValidated decodes the payload of a task and calls registrationValidatedFunc, which handles a user
// who validated their registration. Nothing happens if the function is nil.
func RegistrationValidated(payload []byte, registrationValidatedFunc func(ctx context.Context, userID string)) {
	if registrationValidatedFunc == nil {
		return
	}

	p := queue.TaskRegistrationValidated{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	registrationValidatedFunc(context.Background(), p.UserID)
}
