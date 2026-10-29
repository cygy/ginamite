package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// IsValidAuthToken : returns trus if the auth token is valid.
func IsValidAuthToken(c *gin.Context, tokenID string) (bool, error) {
	// Checks if the auth token is valid and not expired/revoked.
	mongoSession := session.Get(c)
	return model.IsValidAuthToken(c, tokenID, mongoSession)
}

// GetUserAndLocaleFromAuthToken : returns the user struct and the user's locale.
func GetUserAndLocaleFromAuthToken(token *jwt.Token) (any, string, error) {
	user, err := model.UserFromAuthToken(token)
	if err != nil {
		return nil, "", err
	}
	return user, user.Settings.Locale, nil
}

// GetUserAbilities ; returns the abilities of an user.
func GetUserAbilities(c *gin.Context, userID string) (bool, []string) {
	user := context.GetFullUser(c)
	return user.IsAdmin, user.Abilities
}

// GetLatestVersionOfTermsAcceptedByUser : returns the latest version of the terms accepted by the user.
func GetLatestVersionOfTermsAcceptedByUser(c *gin.Context, userID string) string {
	user := context.GetFullUser(c)
	return user.VersionOfTermsAccepted
}
