package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// DoesAccountWithUsernameExist : returns true if an account with this username exists.
func DoesAccountWithUsernameExist(c *gin.Context, username string) bool {
	mongoSession := session.Get(c)
	ok := model.IsUserWithUsernameExisting(username, mongoSession)

	if !ok {
		ok = model.IsDisabledUserWithUsernameExisting(username, mongoSession)
	}

	return ok
}

// DoesAccountWithEmailAddressExist : returns true if an account with this email address exists.
func DoesAccountWithEmailAddressExist(c *gin.Context, emailAddress string) bool {
	mongoSession := session.Get(c)
	ok := model.IsUserWithEmailAddressExisting(emailAddress, mongoSession)

	if !ok {
		ok = model.IsDisabledUserWithEmailAddressExisting(emailAddress, mongoSession)
	}

	return ok
}

// DoesUserWithFacebookUserIDExist : returns true if an account with this facebook id exists.
func DoesUserWithFacebookUserIDExist(c *gin.Context, userID string) bool {
	mongoSession := session.Get(c)
	ok := model.IsUserWithFacebookUserIDExisting(userID, mongoSession)

	if !ok {
		ok = model.IsDisabledUserWithFacebookUserIDExisting(userID, mongoSession)
	}

	return ok
}

// DoesUserWithGoogleUserIDExist : returns true if an account with this google id exists.
func DoesUserWithGoogleUserIDExist(c *gin.Context, userID string) bool {
	mongoSession := session.Get(c)
	ok := model.IsUserWithGoogleUserIDExisting(userID, mongoSession)

	if !ok {
		ok = model.IsDisabledUserWithGoogleUserIDExisting(userID, mongoSession)
	}

	return ok
}
