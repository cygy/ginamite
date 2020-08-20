package user

import "github.com/gin-gonic/gin"

// GetUsersForAdminFunc : returns the list of the users to display in the admin panel.
type GetUsersForAdminFunc func(c *gin.Context) (interface{}, error)

// UpdatePasswordByAdminFunc : updates the password  of a user by a admin.
type UpdatePasswordByAdminFunc func(c *gin.Context, userID, password string) error

// UpdateInformationsByAdminFunc : updates the informations of a user by a admin.
type UpdateInformationsByAdminFunc func(c *gin.Context, userID, firstName, lastName, emailAddress string, validEmailAddress bool) error

// UpdateDescriptionByAdminFunc : updates the description of a user by a admin.
type UpdateDescriptionByAdminFunc func(c *gin.Context, userID, locale, description string) error

// UpdateSocialByAdminFunc : updates the social networks of a user by a admin.
type UpdateSocialByAdminFunc func(c *gin.Context, userID, facebook, twitter, instagram string) error

// UpdateSettingsByAdminFunc : updates the settings of a user by a admin.
type UpdateSettingsByAdminFunc func(c *gin.Context, userID, locale, timezone string) error

// UpdateNotificationsByAdminFunc : updates the notifications of a user by a admin.
type UpdateNotificationsByAdminFunc func(c *gin.Context, userID string, notifications []string) error

// GetTokensForAdminFunc : returns the list of the user's auth tokens to display in the admin panel.
type GetTokensForAdminFunc func(c *gin.Context, userID string) (interface{}, error)

// DeleteTokenByIDByAdminFunc : deletes a user's auth token by an admin.
type DeleteTokenByIDByAdminFunc func(c *gin.Context, tokenID string) error

// SetAbilitiesToUserFunc : sets some abilities to a user.
type SetAbilitiesToUserFunc func(c *gin.Context, abilities []string, userID string) error

// AddAbilitiesToUserFunc : add some abilities to a user.
type AddAbilitiesToUserFunc func(c *gin.Context, abilities []string, userID string) error

// RemoveAbilitiesFromUserFunc : remove some abilities from a user.
type RemoveAbilitiesFromUserFunc func(c *gin.Context, abilities []string, userID string) error
