package functions

import (
	"time"

	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/notifications"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/globalsign/mgo"
)

// DeleteUserByID : deletes an user and all its bound documents.
func DeleteUserByID(updateDocumentsReferencingToDeletedUser func(userID string, mongoSession *mgo.Database)) func(userID string) map[string]time.Time {
	return func(userID string) map[string]time.Time {
		session, db := session.Copy()
		defer session.Close()

		if updateDocumentsReferencingToDeletedUser != nil {
			updateDocumentsReferencingToDeletedUser(userID, db)
		}

		model.DeleteDeletedUserByUserID(userID, db)
		model.DeleteUserByUserID(userID, db)

		authTokens, err := model.GetAuthTokens(userID, db)
		revokedAuthTokens := make(map[string]time.Time, len(authTokens))
		if err == nil {
			for _, authToken := range authTokens {
				revokedAuthTokens[authToken.ID.Hex()] = authToken.ExpiresAt
			}
		}

		model.DeleteAuthTokensByUser(userID, db)

		return revokedAuthTokens
	}
}

// DisableUserByID : disables an user and all its bound documents.
func DisableUserByID(updateDocumentsReferencingToDisabledUser func(userID string, mongoSession *mgo.Database)) func(userID string) map[string]time.Time {
	return func(userID string) map[string]time.Time {
		session, db := session.Copy()
		defer session.Close()

		if updateDocumentsReferencingToDisabledUser != nil {
			updateDocumentsReferencingToDisabledUser(userID, db)
		}

		model.DeleteUserByUserID(userID, db)

		authTokens, err := model.GetAuthTokens(userID, db)
		revokedAuthTokens := make(map[string]time.Time, len(authTokens))
		if err == nil {
			for _, authToken := range authTokens {
				revokedAuthTokens[authToken.ID.Hex()] = authToken.ExpiresAt
			}
		}

		model.DeleteAuthTokensByUser(userID, db)

		return revokedAuthTokens
	}
}

// EnableUserByID : enables an user and all its bound documents.
func EnableUserByID(updateDocumentsReferencingToEnabledUser func(userID string, mongoSession *mgo.Database)) func(userID string) {
	return func(userID string) {
		session, db := session.Copy()
		defer session.Close()

		if updateDocumentsReferencingToEnabledUser != nil {
			updateDocumentsReferencingToEnabledUser(userID, db)
		}

		model.DeleteDisabledUserByUserID(userID, db)
	}
}

// UpdateUserSocialNetworksByID : updates the user's social networks to its bound documents.
func UpdateUserSocialNetworksByID(UpdateDocumentsReferencingToUpdatedUserSocialNetworks func(userID string, mongoSession *mgo.Database)) func(userID string) {
	return func(userID string) {
		if UpdateDocumentsReferencingToUpdatedUserSocialNetworks != nil {
			session, db := session.Copy()
			defer session.Close()

			UpdateDocumentsReferencingToUpdatedUserSocialNetworks(userID, db)
		}
	}

}

// GetNotificationTargetsByUserAndType : gets the targets of a notification sent to an user.
func GetNotificationTargetsByUserAndType(userID, notificationType string) []notifications.Target {
	session, db := session.Copy()
	defer session.Close()

	targets := model.GetUserNotificationTargetsByType(userID, notificationType, db)

	notificationsTargets := make([]notifications.Target, len(targets))
	for i, target := range targets {
		notificationsTargets[i] = notifications.Target{
			Target: target.Type,
			Value:  target.Value,
		}
	}

	return notificationsTargets
}

// IsEmailAddressValid : returns true if the email address is valid and messages can be sent it.
func IsEmailAddressValid(emailAddress string) bool {
	session, db := session.Copy()
	defer session.Close()

	ok := model.IsUserWithValidEmailAddressExisting(emailAddress, db)
	if !ok {
		ok = model.IsDisabledUserWithValidEmailAddressExisting(emailAddress, db)
	}

	return ok
}

// GetNeverUsedAccounts : returns the count of unused accounts deleted (never logged in).
func GetNeverUsedAccounts(intervalInDays uint) []string {
	session, db := session.Copy()
	defer session.Close()

	userIDs, _ := model.GetNeverUsedUserIDs(24*time.Hour*time.Duration(intervalInDays), db)
	return userIDs
}

// GetInactiveAccounts : returns the count of inactive accounts deleted.
func GetInactiveAccounts(intervalInMonths uint) []string {
	session, db := session.Copy()
	defer session.Close()

	userIDs, _ := model.GetInactiveUserIDs(30*24*time.Hour*time.Duration(intervalInMonths), db)
	return userIDs
}

// IsEmptyUser : returns true if the user has never done at least one action.
func IsEmptyUser(isEmptyUser func(userId string, mongoSession *mgo.Database) bool) func(userID string) bool {
	return func(userID string) bool {
		if isEmptyUser == nil {
			return true
		}

		session, db := session.Copy()
		defer session.Close()

		return isEmptyUser(userID, db)
	}
}
