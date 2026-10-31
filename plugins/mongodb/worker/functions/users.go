package functions

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

// SanitizeAccounts returns the maintenance function that runs sanitize, which updates or deletes the data left by
// the deleted users, and returns the number of documents sanitized. It returns 0 if sanitize is nil.
func (f *Functions) SanitizeAccounts(sanitize func(mongoSession store.Database) uint) func() uint {
	return func() uint {
		if sanitize == nil {
			return 0
		}

		return sanitize(f.database())
	}
}

// DeleteUserByID returns the function that deletes a user and all its bound documents: the documents referencing the
// user are updated by the given function, if any, then the user, its deletion record and its auth tokens are deleted.
// The returned function gives the expiration date of each revoked auth token, by token ID.
func (f *Functions) DeleteUserByID(updateDocumentsReferencingToDeletedUser func(ctx context.Context, userID string, mongoSession store.Database)) func(ctx context.Context, userID string) map[string]time.Time {
	return func(ctx context.Context, userID string) map[string]time.Time {
		db := f.database()

		if updateDocumentsReferencingToDeletedUser != nil {
			updateDocumentsReferencingToDeletedUser(ctx, userID, db)
		}

		if err := model.DeleteDeletedUserByUserID(ctx, userID, db); err != nil {
			log.WithField("error", err).Error("unable to delete the deleted user")
		}
		if err := model.DeleteUserByUserID(ctx, userID, db); err != nil {
			log.WithField("error", err).Error("unable to delete the user")
		}

		return revokeAuthTokens(ctx, userID, db)
	}
}

// DisableUserByID returns the function that disables a user and all its bound documents: the documents referencing the
// user are updated by the given function, if any, then the user and its auth tokens are deleted. The disabled user is
// saved by the API. The returned function gives the expiration date of each revoked auth token, by token ID.
func (f *Functions) DisableUserByID(updateDocumentsReferencingToDisabledUser func(ctx context.Context, userID string, mongoSession store.Database)) func(ctx context.Context, userID string) map[string]time.Time {
	return func(ctx context.Context, userID string) map[string]time.Time {
		db := f.database()

		if updateDocumentsReferencingToDisabledUser != nil {
			updateDocumentsReferencingToDisabledUser(ctx, userID, db)
		}

		if err := model.DeleteUserByUserID(ctx, userID, db); err != nil {
			log.WithField("error", err).Error("unable to delete the user")
		}

		return revokeAuthTokens(ctx, userID, db)
	}
}

// EnableUserByID returns the function that enables a user and all its bound documents: the documents referencing the
// user are updated by the given function, if any, then the disabled user is deleted. The enabled user is saved by the
// API.
func (f *Functions) EnableUserByID(updateDocumentsReferencingToEnabledUser func(ctx context.Context, userID string, mongoSession store.Database)) func(ctx context.Context, userID string) {
	return func(ctx context.Context, userID string) {
		db := f.database()

		if updateDocumentsReferencingToEnabledUser != nil {
			updateDocumentsReferencingToEnabledUser(ctx, userID, db)
		}

		if err := model.DeleteDisabledUserByUserID(ctx, userID, db); err != nil {
			log.WithField("error", err).Error("unable to delete the disabled user")
		}
	}
}

// UpdateUserSocialNetworksByID returns the function that runs the given update of the documents referencing the
// social networks of a user. It does nothing if the update is nil.
func (f *Functions) UpdateUserSocialNetworksByID(updateDocumentsReferencingToUpdatedUserSocialNetworks func(ctx context.Context, userID string, mongoSession store.Database)) func(ctx context.Context, userID string) {
	return func(ctx context.Context, userID string) {
		if updateDocumentsReferencingToUpdatedUserSocialNetworks == nil {
			return
		}

		updateDocumentsReferencingToUpdatedUserSocialNetworks(ctx, userID, f.database())
	}
}

// GetNotificationTargetsByUserAndType returns the targets of a notification sent to a user: the devices and the email
// address, see model.GetUserNotificationTargetsByType.
func (f *Functions) GetNotificationTargetsByUserAndType(ctx context.Context, userID, notificationType string) ([]notifications.Target, error) {
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

// IsEmailAddressValid reports whether a user, enabled or disabled, has this email address and it is valid, i.e. messages
// can be sent to it.
func (f *Functions) IsEmailAddressValid(ctx context.Context, emailAddress string) (bool, error) {
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

// GetNeverUsedAccounts returns the IDs of the accounts that never signed in, and registered more than
// intervalInDays days ago. A failing query is logged and gives no account.
func (f *Functions) GetNeverUsedAccounts(ctx context.Context, intervalInDays uint) []string {
	db := f.database()
	userIDs, err := model.GetNeverUsedUserIDs(ctx, 24*time.Hour*time.Duration(intervalInDays), db)
	if err != nil {
		log.WithField("error", err).Error("unable to get the accounts never used")
	}

	return userIDs
}

// GetInactiveAccounts returns the IDs of the accounts whose last sign-in is older than intervalInMonths months (of 30
// days). A failing query is logged and gives no account.
func (f *Functions) GetInactiveAccounts(ctx context.Context, intervalInMonths uint) []string {
	db := f.database()
	userIDs, err := model.GetInactiveUserIDs(ctx, 30*24*time.Hour*time.Duration(intervalInMonths), db)
	if err != nil {
		log.WithField("error", err).Error("unable to get the inactive accounts")
	}

	return userIDs
}

// IsEmptyUser returns the function that reports whether a user has never done any action, according to isEmptyUser.
// Every user is empty if isEmptyUser is nil.
func (f *Functions) IsEmptyUser(isEmptyUser func(ctx context.Context, userID string, mongoSession store.Database) bool) func(ctx context.Context, userID string) bool {
	return func(ctx context.Context, userID string) bool {
		if isEmptyUser == nil {
			return true
		}

		db := f.database()
		return isEmptyUser(ctx, userID, db)
	}
}

// revokeAuthTokens deletes the auth tokens of the user and returns the expiration date of each, by token ID.
func revokeAuthTokens(ctx context.Context, userID string, db store.Database) map[string]time.Time {
	authTokens, err := model.GetAuthTokens(ctx, userID, db)
	revokedAuthTokens := make(map[string]time.Time, len(authTokens))
	if err == nil {
		for _, authToken := range authTokens {
			revokedAuthTokens[authToken.ID.Hex()] = authToken.ExpiresAt
		}
	}

	if _, err := model.DeleteAuthTokensByUser(ctx, userID, db); err != nil {
		log.WithField("error", err).Error("unable to delete the auth tokens of the user")
	}

	return revokedAuthTokens
}
