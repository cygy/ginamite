package functions

import (
	"context"
	"errors"
	"fmt"
	"time"

	errs "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// documentMissing reports whether err says a document does not exist.
func documentMissing(err error) bool {
	return errors.Is(err, mongo.ErrNoDocuments) || errs.IsNotFound(err)
}

// wrapHook adds the context of the hooks to their error, and returns nil if there is none.
func wrapHook(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("unable to update the documents referencing the user: %w", err)
}

// Sanitize runs the hook SanitizeAccounts, which updates or deletes the data left by the deleted users, and returns the
// number of documents sanitized. It returns 0 if there is no hook.
func (f *Functions) Sanitize(_ context.Context) (uint, error) {
	if f.Hooks.SanitizeAccounts == nil {
		return 0, nil
	}

	return f.Hooks.SanitizeAccounts(f.database()), nil
}

// Delete deletes a user and all its bound documents: the documents referencing the user are updated by the hook, if
// any, then the user, its deletion record and its auth tokens are deleted. It gives the expiration date of each revoked
// auth token, by token ID, and the error of the database, if any: it can be called again, a user already deleted is not
// an error.
func (f *Functions) Delete(ctx context.Context, userID string) (map[string]time.Time, error) {
	db := f.database()

	if err := f.runHook(ctx, f.Hooks.UpdateDocumentsReferencingToDeletedUser, userID, db); err != nil {
		return nil, err
	}

	if err := model.DeleteDeletedUserByUserID(ctx, userID, db); err != nil {
		return nil, fmt.Errorf("unable to delete the deleted user: %w", err)
	}
	if err := model.DeleteUserByUserID(ctx, userID, db); err != nil && !documentMissing(err) {
		return nil, fmt.Errorf("unable to delete the user: %w", err)
	}

	return revokeAuthTokens(ctx, userID, db)
}

// Disable disables a user and all its bound documents: the documents referencing the user are updated by the hook, if
// any, then the user and its auth tokens are deleted. The disabled user is saved by the API. It gives the expiration
// date of each revoked auth token, by token ID, and the error of the database, if any: it can be called again, a user
// already deleted is not an error.
func (f *Functions) Disable(ctx context.Context, userID string) (map[string]time.Time, error) {
	db := f.database()

	if err := f.runHook(ctx, f.Hooks.UpdateDocumentsReferencingToDisabledUser, userID, db); err != nil {
		return nil, err
	}

	if err := model.DeleteUserByUserID(ctx, userID, db); err != nil && !documentMissing(err) {
		return nil, fmt.Errorf("unable to delete the user: %w", err)
	}

	return revokeAuthTokens(ctx, userID, db)
}

// Enable enables a user and all its bound documents: the documents referencing the user are updated by the hook, if
// any, then the disabled user is deleted. The enabled user is saved by the API.
func (f *Functions) Enable(ctx context.Context, userID string) error {
	db := f.database()

	if err := f.runHook(ctx, f.Hooks.UpdateDocumentsReferencingToEnabledUser, userID, db); err != nil {
		return err
	}

	if err := model.DeleteDisabledUserByUserID(ctx, userID, db); err != nil {
		return fmt.Errorf("unable to delete the disabled user: %w", err)
	}

	return nil
}

// UpdateSocialNetworks runs the hook updating the documents referencing the social networks of a user. It does nothing
// if there is no hook.
func (f *Functions) UpdateSocialNetworks(ctx context.Context, userID string) error {
	return f.runHook(ctx, f.Hooks.UpdateDocumentsReferencingToUpdatedUserSocialNetworks, userID, f.database())
}

// runHook runs a hook, if any, and adds its context to its error.
func (f *Functions) runHook(ctx context.Context, hook UserHook, userID string, db store.Database) error {
	if hook == nil {
		return nil
	}

	return wrapHook(hook(ctx, userID, db))
}

// Targets returns the targets of a notification sent to a user: the devices and the email
// address, see model.GetUserNotificationTargetsByType.
func (f *Functions) Targets(ctx context.Context, userID, notificationType string) ([]notifications.Target, error) {
	db := f.database()

	targets, err := model.GetUserNotificationTargetsByType(ctx, userID, notificationType, db)
	if err != nil {
		return []notifications.Target{}, err
	}

	notificationsTargets := make([]notifications.Target, len(targets))
	for i, target := range targets {
		notificationsTargets[i] = notifications.Target{
			Target: target.Type,
			Value:  target.Value,
		}
	}

	return notificationsTargets, nil
}

// IsValid reports whether a user, enabled or disabled, has this email address and it is valid, i.e. messages
// can be sent to it.
func (f *Functions) IsValid(ctx context.Context, emailAddress string) (bool, error) {
	db := f.database()

	ok, err := model.IsUserWithValidEmailAddressExisting(ctx, emailAddress, db)
	if err != nil {
		return false, err
	}
	if !ok {
		ok, err = model.IsDisabledUserWithValidEmailAddressExisting(ctx, emailAddress, db)
	}

	return ok, err
}

// NeverUsed returns the IDs of the accounts that never signed in, and registered more than intervalInDays days ago.
func (f *Functions) NeverUsed(ctx context.Context, intervalInDays uint) ([]string, error) {
	userIDs, err := model.GetNeverUsedUserIDs(ctx, 24*time.Hour*time.Duration(intervalInDays), f.database()) //nolint:gosec // an interval in days is far below the overflow
	if err != nil {
		return nil, fmt.Errorf("unable to get the accounts never used: %w", err)
	}

	return userIDs, nil
}

// Inactive returns the IDs of the accounts whose last sign-in is older than intervalInMonths months (of 30 days).
func (f *Functions) Inactive(ctx context.Context, intervalInMonths uint) ([]string, error) {
	userIDs, err := model.GetInactiveUserIDs(ctx, 30*24*time.Hour*time.Duration(intervalInMonths), f.database()) //nolint:gosec // an interval in months is far below the overflow
	if err != nil {
		return nil, fmt.Errorf("unable to get the inactive accounts: %w", err)
	}

	return userIDs, nil
}

// IsEmpty reports whether a user has never done any action, according to the hook IsUserEmpty. Every user is empty if
// there is no hook.
func (f *Functions) IsEmpty(ctx context.Context, userID string) (bool, error) {
	if f.Hooks.IsUserEmpty == nil {
		return true, nil
	}

	return f.Hooks.IsUserEmpty(ctx, userID, f.database()), nil
}

// revokeAuthTokens deletes the auth tokens of the user and returns the expiration date of each, by token ID. The
// tokens are not deleted if they can not be listed, since their revocation could not be told to the other servers.
func revokeAuthTokens(ctx context.Context, userID string, db store.Database) (map[string]time.Time, error) {
	authTokens, err := model.GetAuthTokens(ctx, userID, db)
	if err != nil {
		return nil, fmt.Errorf("unable to get the auth tokens of the user: %w", err)
	}

	revokedAuthTokens := make(map[string]time.Time, len(authTokens))
	for _, authToken := range authTokens {
		revokedAuthTokens[authToken.ID.Hex()] = authToken.ExpiresAt
	}

	if _, err := model.DeleteAuthTokensByUser(ctx, userID, db); err != nil {
		return nil, fmt.Errorf("unable to delete the auth tokens of the user: %w", err)
	}

	return revokedAuthTokens, nil
}
