package functions

import (
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/session"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// IsValidAuthToken reports whether the auth token is valid: neither expired nor revoked.
func IsValidAuthToken(c *gin.Context, tokenID string) (bool, error) {
	mongoSession := session.Get(c)
	return model.IsValidAuthToken(c, tokenID, mongoSession)
}

// GetUserAndLocaleFromAuthToken returns the model.User built from the claims of the auth token, and the locale of
// the user. The user is stored by value in the context, see context.GetUser.
func GetUserAndLocaleFromAuthToken(token *jwt.Token) (any, string, error) {
	user, err := model.UserFromAuthToken(token)
	if err != nil {
		return nil, "", err
	}

	return *user, user.Settings.Locale, nil
}

// GetUserAbilities reports whether the authenticated user is an administrator, and returns the abilities of the user.
// The user is not an administrator and has no ability if it can not be loaded.
func GetUserAbilities(c *gin.Context, userID string) (bool, []string) {
	user := context.GetFullUser(c)
	if user == nil {
		return false, nil
	}

	return user.IsAdmin, user.Abilities
}

// GetLatestVersionOfTermsAcceptedByUser returns the latest version of the terms accepted by the authenticated user,
// or an empty string if the user can not be loaded.
func GetLatestVersionOfTermsAcceptedByUser(c *gin.Context, userID string) string {
	user := context.GetFullUser(c)
	if user == nil {
		return ""
	}

	return user.VersionOfTermsAccepted
}
