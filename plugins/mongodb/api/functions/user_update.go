package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// UpdateUserPassword : update the password of an user.
func UpdateUserPassword(c *gin.Context, userID, password string, isPasswordEncrypted bool) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdatePassword(c, password, isPasswordEncrypted, mongoSession)
}

// UpdateUserEmailAddress : update the email address of an user.
func UpdateUserEmailAddress(c *gin.Context, userID, emailAddress string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateEmailAddress(c, emailAddress, mongoSession)
}

// UserEmailAddressIsVerified : the email address of an user is verified.
func UserEmailAddressIsVerified(c *gin.Context, userID string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.SetEmailAddressValid(c, mongoSession)
}

// UpdateUserSocialNetworks : update the social network links of an user.
func UpdateUserSocialNetworks(c *gin.Context, userID, facebook, twitter, instagram string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateSocialNetworks(c, facebook, twitter, instagram, mongoSession)
}
