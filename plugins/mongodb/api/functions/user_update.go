package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// UpdateUserPassword updates the password of a user.
func UpdateUserPassword(c *gin.Context, userID, password string, isPasswordEncrypted bool) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdatePassword(c, password, isPasswordEncrypted, mongoSession)
}

// UpdateUserEmailAddress updates the email address of a user.
func UpdateUserEmailAddress(c *gin.Context, userID, emailAddress string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateEmailAddress(c, emailAddress, mongoSession)
}

// UserEmailAddressIsVerified marks the email address of the user as verified.
func UserEmailAddressIsVerified(c *gin.Context, userID string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.SetEmailAddressValid(c, mongoSession)
}

// UpdateUserSocialNetworks updates the social network links of a user.
func UpdateUserSocialNetworks(c *gin.Context, userID, facebook, twitter, instagram string) error {
	mongoSession := session.Get(c)
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateSocialNetworks(c, facebook, twitter, instagram, mongoSession)
}

// UpdateVersionOfTermsAcceptedByUser saves the version of the terms accepted by the authenticated user.
func UpdateVersionOfTermsAcceptedByUser(c *gin.Context, userID, termsVersion string) error {
	mongoSession := session.Get(c)
	user := context.GetUser(c)
	return user.AcceptVersionOfTerms(c, termsVersion, mongoSession)
}
