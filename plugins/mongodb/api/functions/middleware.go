package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	jwt "github.com/golang-jwt/jwt"
)

// IsValidAuthToken : returns trus if the auth token is valid.
func IsValidAuthToken(c *gin.Context, tokenID string) bool {
	if !bson.IsObjectIdHex(tokenID) {
		return false
	}

	// Checks if the auth token is valid and not expired/revoked.
	mongoSession := session.Get(c)
	if !model.IsValidAuthToken(tokenID, mongoSession) {
		return false
	}

	return true
}

// GetUserAndLocaleFromAuthToken : returns the user struct and the user's locale.
func GetUserAndLocaleFromAuthToken(token *jwt.Token) (interface{}, string) {
	user := model.UserFromAuthToken(token)
	return user, user.Settings.Locale
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
