package functions

import (
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/plugins/mongodb/api/ctxsettings"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"

	"github.com/gin-gonic/gin"
)

// GetEnabledUsersForAdmin returns the enabled users for the admin panel.
func GetEnabledUsersForAdmin(c *gin.Context) (any, error) {
	mongoSession := session.Get(c)
	users, err := model.GetEnabledUsersForAdmin(c, mongoSession)

	return users, err
}

// GetDisabledUsersForAdmin returns the disabled users for the admin panel.
func GetDisabledUsersForAdmin(c *gin.Context) (any, error) {
	mongoSession := session.Get(c)
	users, err := model.GetDisabledUsersForAdmin(c, mongoSession)

	return users, err
}

// UpdateUserPasswordByAdmin updates the password of a user by an admin.
func UpdateUserPasswordByAdmin(c *gin.Context, userID, password string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdatePassword(c, password, false, mongoSession)
}

// UpdateUserInformationsByAdmin updates the informations of a user by an admin.
func UpdateUserInformationsByAdmin(c *gin.Context, userID, firstName, lastName, emailAddress string, validEmailAddress bool) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateAllInformations(c, firstName, lastName, emailAddress, validEmailAddress, mongoSession)
}

// UpdateUserDescriptionByAdmin updates the description of a user by an admin.
func UpdateUserDescriptionByAdmin(c *gin.Context, userID, locale, description string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdatePublicProfile(c, locale, description, mongoSession)
}

// UpdateUserSocialByAdmin updates the social networks of a user by an admin.
func UpdateUserSocialByAdmin(c *gin.Context, userID, facebook, twitter, instagram string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateSocialNetworks(c, facebook, twitter, instagram, mongoSession)
}

// UpdateUserSettingsByAdmin updates the settings of a user by an admin.
func UpdateUserSettingsByAdmin(c *gin.Context, userID, locale, timezone string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateSettings(c, locale, timezone, mongoSession)
}

// UpdateUserNotificationsByAdmin updates the notifications of a user by an admin.
func UpdateUserNotificationsByAdmin(c *gin.Context, userID string, notifications []string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateNotifications(c, notifications, mongoSession)
}

// GetTokensForAdmin returns the auth tokens of the user, as model.AdminAuthToken, for the admin panel.
func GetTokensForAdmin(c *gin.Context, userID string) (any, error) {
	mongoSession := session.Get(c)

	tokens, err := model.GetAuthTokens(c, userID, mongoSession)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	adminTokens := make([]*model.AdminAuthToken, len(tokens))
	for i, token := range tokens {
		adminTokens[i] = model.NewAdminAuthToken(&token)
	}

	return adminTokens, nil
}

// DeleteTokenByIDByAdmin deletes a user's auth token by an admin.
func DeleteTokenByIDByAdmin(c *gin.Context, tokenID string) error {
	mongoSession := session.Get(c)
	return model.DeleteAuthTokenByID(c, tokenID, mongoSession)
}

// SetAbilitiesToUser sets the abilities to a user.
func SetAbilitiesToUser(c *gin.Context, abilities []string, userID string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.SetAbilities(c, abilities, ctxsettings.GetSettings(c).AllUserAbilities, mongoSession)
}

// AddAbilitiesToUser adds some abilities to a user.
func AddAbilitiesToUser(c *gin.Context, abilities []string, userID string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.AddAbilities(c, abilities, ctxsettings.GetSettings(c).AllUserAbilities, mongoSession)
}

// RemoveAbilitiesFromUser removes some abilities from a user.
func RemoveAbilitiesFromUser(c *gin.Context, abilities []string, userID string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.RemoveAbilities(c, abilities, mongoSession)
}
