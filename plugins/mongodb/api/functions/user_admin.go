package functions

import (
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// GetEnabledUsersForAdmin : returns the enabled users for the admin panel.
func GetEnabledUsersForAdmin(c *gin.Context) (interface{}, error) {
	mongoSession := session.Get(c)
	users, err := model.GetEnabledUsersForAdmin(mongoSession)

	return users, err
}

// GetDisabledUsersForAdmin : returns the disabled users for the admin panel.
func GetDisabledUsersForAdmin(c *gin.Context) (interface{}, error) {
	mongoSession := session.Get(c)
	users, err := model.GetDisabledUsersForAdmin(mongoSession)

	return users, err
}

// UpdateUserPasswordByAdmin : updates the password of a user by an admin.
func UpdateUserPasswordByAdmin(c *gin.Context, userID, password string) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.UpdatePassword(password, false, mongoSession)
}

// UpdateUserInformationsByAdmin : updates the informations of a user by an admin.
func UpdateUserInformationsByAdmin(c *gin.Context, userID, firstName, lastName, emailAddress string, validEmailAddress bool) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.UpdateAllInformations(firstName, lastName, emailAddress, validEmailAddress, mongoSession)
}

// UpdateUserDescriptionByAdmin : updates the description of a user by an admin.
func UpdateUserDescriptionByAdmin(c *gin.Context, userID, locale, description string) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.UpdatePublicProfile(locale, description, mongoSession)
}

// UpdateUserSocialByAdmin : updates the social networks of a user by an admin.
func UpdateUserSocialByAdmin(c *gin.Context, userID, facebook, twitter, instagram string) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.UpdateSocialNetworks(facebook, twitter, instagram, mongoSession)
}

// UpdateUserSettingsByAdmin : updates the settings of a user by an admin.
func UpdateUserSettingsByAdmin(c *gin.Context, userID, locale, timezone string) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.UpdateSettings(locale, timezone, mongoSession)
}

// UpdateUserNotificationsByAdmin : updates the notifications of a user by an admin.
func UpdateUserNotificationsByAdmin(c *gin.Context, userID string, notifications []string) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.UpdateNotifications(notifications, mongoSession)
}

// GetTokensForAdmin : returns the user's auth tokens for the admin panel.
func GetTokensForAdmin(c *gin.Context, userID string) (interface{}, error) {
	mongoSession := session.Get(c)

	tokens, err := model.GetAuthTokens(userID, mongoSession)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	adminTokens := make([]*model.AdminAuthToken, len(tokens))
	for i, token := range tokens {
		adminTokens[i] = model.NewAdminAuthToken(&token)
	}

	return adminTokens, nil
}

// DeleteTokenByIDByAdmin : deletes a user's auth token by an admin.
func DeleteTokenByIDByAdmin(c *gin.Context, tokenID string) error {
	mongoSession := session.Get(c)
	return model.DeleteAuthTokenByID(tokenID, mongoSession)
}

// SetAbilitiesToUser : sets the abilities to an user.
func SetAbilitiesToUser(c *gin.Context, abilities []string, userID string) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.SetAbilities(abilities, mongoSession)
}

// AddAbilitiesToUser : adds some abilities to an user.
func AddAbilitiesToUser(c *gin.Context, abilities []string, userID string) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.AddAbilities(abilities, mongoSession)
}

// RemoveAbilitiesFromUser : remove some abilities from an user.
func RemoveAbilitiesFromUser(c *gin.Context, abilities []string, userID string) error {
	mongoSession := session.Get(c)
	user := model.UserWithID(userID)

	return user.RemoveAbilities(abilities, mongoSession)
}
