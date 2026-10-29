package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// DoesAccountWithUsernameExist : returns true if an account with this username exists.
func DoesAccountWithUsernameExist(c *gin.Context, username string) (bool, error) {
	mongoSession := session.Get(c)
	ok, err := model.IsUserWithUsernameExisting(c, username, mongoSession)
	if err != nil {
		return false, err
	}

	if !ok {
		ok, err = model.IsDisabledUserWithUsernameExisting(c, username, mongoSession)
	}

	return ok, err
}

// DoesAccountWithEmailAddressExist : returns true if an account with this email address exists.
func DoesAccountWithEmailAddressExist(c *gin.Context, emailAddress string) (bool, error) {
	mongoSession := session.Get(c)
	ok, err := model.IsUserWithEmailAddressExisting(c, emailAddress, mongoSession)
	if err != nil {
		return false, err
	}

	if !ok {
		ok, err = model.IsDisabledUserWithEmailAddressExisting(c, emailAddress, mongoSession)
	}

	return ok, err
}

// DoesUserWithFacebookUserIDExist : returns true if an account with this facebook id exists.
func DoesUserWithFacebookUserIDExist(c *gin.Context, userID string) (bool, error) {
	mongoSession := session.Get(c)
	ok, err := model.IsUserWithFacebookUserIDExisting(c, userID, mongoSession)
	if err != nil {
		return false, err
	}

	if !ok {
		ok, err = model.IsDisabledUserWithFacebookUserIDExisting(c, userID, mongoSession)
	}

	return ok, err
}

// DoesUserWithGoogleUserIDExist : returns true if an account with this google id exists.
func DoesUserWithGoogleUserIDExist(c *gin.Context, userID string) (bool, error) {
	mongoSession := session.Get(c)
	ok, err := model.IsUserWithGoogleUserIDExisting(c, userID, mongoSession)
	if err != nil {
		return false, err
	}

	if !ok {
		ok, err = model.IsDisabledUserWithGoogleUserIDExisting(c, userID, mongoSession)
	}

	return ok, err
}
