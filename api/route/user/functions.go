package user

import "github.com/gin-gonic/gin"

// functionsKey is the key of the functions in the Gin context.
const functionsKey = "_user_functions"

// Functions are the functions provided by the application, such as the ones of the storage plugin, used by the
// handlers of the package. A missing function disables the feature relying on it.
type Functions struct {
	// GetEnabledUsersForAdmin is a function to return the list of the users to display in the admin panel.
	GetEnabledUsersForAdmin GetUsersForAdminFunc

	// GetDisabledUsersForAdmin is a function to return the list of the disabled users to display in the admin panel.
	GetDisabledUsersForAdmin GetUsersForAdminFunc

	// UpdatePasswordByAdmin is a function to update the password of a user by a admin.
	UpdatePasswordByAdmin UpdatePasswordByAdminFunc

	// UpdateInformationsByAdmin is a function to update the informations of a user by a admin.
	UpdateInformationsByAdmin UpdateInformationsByAdminFunc

	// UpdateDescriptionByAdmin is a function to update the description of a user by a admin.
	UpdateDescriptionByAdmin UpdateDescriptionByAdminFunc

	// UpdateSocialByAdmin is a function to update the social networks of a user by a admin.
	UpdateSocialByAdmin UpdateSocialByAdminFunc

	// UpdateSettingsByAdmin is a function to update the informations of a user by a admin.
	UpdateSettingsByAdmin UpdateSettingsByAdminFunc

	// UpdateNotificationsByAdmin is a function to update the notifications of a user by a admin.
	UpdateNotificationsByAdmin UpdateNotificationsByAdminFunc

	// GetTokensForAdmin is a function to return the list of the user's auth tokens to display in the admin panel.
	GetTokensForAdmin GetTokensForAdminFunc

	// DeleteTokenByIDByAdmin is a function to delete a user's auth tokens by an admin.
	DeleteTokenByIDByAdmin DeleteTokenByIDByAdminFunc

	// SetAbilitiesToUser sets abilities to a user.
	SetAbilitiesToUser SetAbilitiesToUserFunc

	// AddAbilitiesToUser add some abilities to a user.
	AddAbilitiesToUser AddAbilitiesToUserFunc

	// RemoveAbilitiesFromUser remove some abilities from a user.
	RemoveAbilitiesFromUser RemoveAbilitiesFromUserFunc
}

// Inject returns a middleware saving the functions in the Gin context, see FunctionsOf.
func Inject(functions *Functions) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetFunctions(c, functions)
		c.Next()
	}
}

// SetFunctions saves the functions in the Gin context, like Inject does. It is meant for the tests.
func SetFunctions(c *gin.Context, functions *Functions) {
	c.Set(functionsKey, functions)
}

// FunctionsOf returns the functions saved in the Gin context by Inject.
// It panics if there are none: the handlers can not work without them.
func FunctionsOf(c *gin.Context) *Functions {
	value, ok := c.Get(functionsKey)
	if !ok {
		panic("user: no functions in the context, add the middleware Inject")
	}

	return value.(*Functions)
}
