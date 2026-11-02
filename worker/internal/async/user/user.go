// Package user provides the asynchronous tasks about the users.
//
// The tasks return the error of their function, or a permanent error, see broker.Permanent, if the payload can not be
// decoded.
package user

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/queue"
)

// DeleteByID decodes the payload of a task to delete a user and calls deleteUserByIDFunc, which deletes the user
// and returns the auth tokens revoked, defined by their ID and their expiration date. The tokens are then invalidated.
// Nothing happens if the function is nil.
func DeleteByID(payload []byte, q *queue.Queue, deleteUserByIDFunc func(ctx context.Context, userID string) (map[string]time.Time, error)) error {
	if deleteUserByIDFunc == nil {
		return nil
	}

	p := queue.TaskDeleteUser{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	revokedAuthTokens, err := deleteUserByIDFunc(context.Background(), p.UserID)
	if err != nil {
		return err
	}

	if len(revokedAuthTokens) > 0 {
		q.InvalidAuthTokens(revokedAuthTokens)
	}

	return nil
}

// DisableByID decodes the payload of a task to disable a user and calls disableUserByIDFunc, which disables the user
// and returns the auth tokens revoked, defined by their ID and their expiration date. The tokens are then invalidated.
// Nothing happens if the function is nil.
func DisableByID(payload []byte, q *queue.Queue, disableUserByIDFunc func(ctx context.Context, userID string) (map[string]time.Time, error)) error {
	if disableUserByIDFunc == nil {
		return nil
	}

	p := queue.TaskDisableUser{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	revokedAuthTokens, err := disableUserByIDFunc(context.Background(), p.UserID)
	if err != nil {
		return err
	}

	if len(revokedAuthTokens) > 0 {
		q.InvalidAuthTokens(revokedAuthTokens)
	}

	return nil
}

// EnableByID decodes the payload of a task to enable a user and calls enableUserByIDFunc.
// Nothing happens if the function is nil.
func EnableByID(payload []byte, enableUserByIDFunc func(ctx context.Context, userID string) error) error {
	if enableUserByIDFunc == nil {
		return nil
	}

	p := queue.TaskEnableUser{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	return enableUserByIDFunc(context.Background(), p.UserID)
}

// UpdateSocialNetworksByID decodes the payload of a task and calls updateUserSocialNetworksByIDFunc, which updates
// the documents referencing the social networks of a user. Nothing happens if the function is nil.
func UpdateSocialNetworksByID(payload []byte, updateUserSocialNetworksByIDFunc func(ctx context.Context, userID string) error) error {
	if updateUserSocialNetworksByIDFunc == nil {
		return nil
	}

	p := queue.TaskUpdateUserSocialNetworks{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	return updateUserSocialNetworksByIDFunc(context.Background(), p.UserID)
}

// RegistrationDone decodes the payload of a task and calls registrationDoneFunc, which handles a user who registered.
// Nothing happens if the function is nil.
func RegistrationDone(payload []byte, registrationDoneFunc func(ctx context.Context, userID string) error) error {
	if registrationDoneFunc == nil {
		return nil
	}

	p := queue.TaskRegistrationDone{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	return registrationDoneFunc(context.Background(), p.UserID)
}

// RegistrationValidated decodes the payload of a task and calls registrationValidatedFunc, which handles a user
// who validated their registration. Nothing happens if the function is nil.
func RegistrationValidated(payload []byte, registrationValidatedFunc func(ctx context.Context, userID string) error) error {
	if registrationValidatedFunc == nil {
		return nil
	}

	p := queue.TaskRegistrationValidated{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	return registrationValidatedFunc(context.Background(), p.UserID)
}
