package functions

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SanitizeAccounts : maintenance function to update/delete the data from a deleted user.
func SanitizeAccounts(SanitizeAccounts func(mongoSession *mongo.Database) uint) func() uint {
	return func() uint {
		if SanitizeAccounts == nil {
			return 0
		}
		db := session.Database()
		return SanitizeAccounts(db)
	}
}

// DeleteUserByID : deletes an user and all its bound documents.
func DeleteUserByID(updateDocumentsReferencingToDeletedUser func(ctx context.Context, userID string, mongoSession *mongo.Database)) func(ctx context.Context, userID string) map[string]time.Time {
	return func(ctx context.Context, userID string) map[string]time.Time {
		db := session.Database()

		if updateDocumentsReferencingToDeletedUser != nil {
			updateDocumentsReferencingToDeletedUser(ctx, userID, db)
		}

		model.DeleteDeletedUserByUserID(ctx, userID, db)
		model.DeleteUserByUserID(ctx, userID, db)

		authTokens, err := model.GetAuthTokens(ctx, userID, db)
		revokedAuthTokens := make(map[string]time.Time, len(authTokens))
		if err == nil {
			for _, authToken := range authTokens {
				revokedAuthTokens[authToken.ID.Hex()] = authToken.ExpiresAt
			}
		}

		model.DeleteAuthTokensByUser(ctx, userID, db)

		return revokedAuthTokens
	}
}

// DisableUserByID : disables an user and all its bound documents.
func DisableUserByID(updateDocumentsReferencingToDisabledUser func(ctx context.Context, userID string, mongoSession *mongo.Database)) func(ctx context.Context, userID string) map[string]time.Time {
	return func(ctx context.Context, userID string) map[string]time.Time {
		db := session.Database()

		if updateDocumentsReferencingToDisabledUser != nil {
			updateDocumentsReferencingToDisabledUser(ctx, userID, db)
		}

		model.DeleteUserByUserID(ctx, userID, db)

		authTokens, err := model.GetAuthTokens(ctx, userID, db)
		revokedAuthTokens := make(map[string]time.Time, len(authTokens))
		if err == nil {
			for _, authToken := range authTokens {
				revokedAuthTokens[authToken.ID.Hex()] = authToken.ExpiresAt
			}
		}

		model.DeleteAuthTokensByUser(ctx, userID, db)

		return revokedAuthTokens
	}
}

// EnableUserByID : enables an user and all its bound documents.
func EnableUserByID(updateDocumentsReferencingToEnabledUser func(ctx context.Context, userID string, mongoSession *mongo.Database)) func(ctx context.Context, userID string) {
	return func(ctx context.Context, userID string) {
		db := session.Database()

		if updateDocumentsReferencingToEnabledUser != nil {
			updateDocumentsReferencingToEnabledUser(ctx, userID, db)
		}

		model.DeleteDisabledUserByUserID(ctx, userID, db)
	}
}

// UpdateUserSocialNetworksByID : updates the user's social networks to its bound documents.
func UpdateUserSocialNetworksByID(UpdateDocumentsReferencingToUpdatedUserSocialNetworks func(ctx context.Context, userID string, mongoSession *mongo.Database)) func(ctx context.Context, userID string) {
	return func(ctx context.Context, userID string) {
		if UpdateDocumentsReferencingToUpdatedUserSocialNetworks == nil {
			return
		}
		db := session.Database()
		UpdateDocumentsReferencingToUpdatedUserSocialNetworks(ctx, userID, db)
	}

}

// GetNotificationTargetsByUserAndType : gets the targets of a notification sent to an user.
func GetNotificationTargetsByUserAndType(ctx context.Context, userID, notificationType string) ([]notifications.Target, error) {
	db := session.Database()

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

// IsEmailAddressValid : returns true if the email address is valid and messages can be sent it.
func IsEmailAddressValid(ctx context.Context, emailAddress string) (bool, error) {
	db := session.Database()

	ok, err := model.IsUserWithValidEmailAddressExisting(ctx, emailAddress, db)
	if err != nil {
		return false, err
	}
	if !ok {
		ok, err = model.IsDisabledUserWithValidEmailAddressExisting(ctx, emailAddress, db)
	}

	return ok, err
}

// GetNeverUsedAccounts : returns the count of unused accounts deleted (never logged in).
func GetNeverUsedAccounts(ctx context.Context, intervalInDays uint) []string {
	db := session.Database()
	userIDs, _ := model.GetNeverUsedUserIDs(ctx, 24*time.Hour*time.Duration(intervalInDays), db)
	return userIDs
}

// GetInactiveAccounts : returns the count of inactive accounts deleted.
func GetInactiveAccounts(ctx context.Context, intervalInMonths uint) []string {
	db := session.Database()
	userIDs, _ := model.GetInactiveUserIDs(ctx, 30*24*time.Hour*time.Duration(intervalInMonths), db)
	return userIDs
}

// IsEmptyUser : returns true if the user has never done at least one action.
func IsEmptyUser(isEmptyUser func(ctx context.Context, userId string, mongoSession *mongo.Database) bool) func(ctx context.Context, userID string) bool {
	return func(ctx context.Context, userID string) bool {
		if isEmptyUser == nil {
			return true
		}

		db := session.Database()
		return isEmptyUser(ctx, userID, db)
	}
}
