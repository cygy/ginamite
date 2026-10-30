package functions

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
)

// DoesAccountWithUsernameExist reports whether an account with this username exists.
func DoesAccountWithUsernameExist(c *gin.Context, username string) (bool, error) {
	return existsInEither(c, username, model.IsUserWithUsernameExisting, model.IsDisabledUserWithUsernameExisting)
}

// DoesAccountWithEmailAddressExist reports whether an account with this email address exists.
func DoesAccountWithEmailAddressExist(c *gin.Context, emailAddress string) (bool, error) {
	return existsInEither(c, emailAddress, model.IsUserWithEmailAddressExisting, model.IsDisabledUserWithEmailAddressExisting)
}

// DoesUserWithFacebookUserIDExist reports whether an account with this facebook id exists.
func DoesUserWithFacebookUserIDExist(c *gin.Context, userID string) (bool, error) {
	return existsInEither(c, userID, model.IsUserWithFacebookUserIDExisting, model.IsDisabledUserWithFacebookUserIDExisting)
}

// DoesUserWithGoogleUserIDExist reports whether an account with this google id exists.
func DoesUserWithGoogleUserIDExist(c *gin.Context, userID string) (bool, error) {
	return existsInEither(c, userID, model.IsUserWithGoogleUserIDExisting, model.IsDisabledUserWithGoogleUserIDExisting)
}

// existsInEither reports whether the value is found by the check of the enabled users or, if not, by the check of
// the disabled users.
func existsInEither(c *gin.Context, value string, enabled, disabled func(context.Context, string, store.Database) (bool, error)) (bool, error) {
	mongoSession := session.Get(c)

	ok, err := enabled(c, value, mongoSession)
	if err != nil {
		return false, err
	}

	if !ok {
		ok, err = disabled(c, value, mongoSession)
	}

	return ok, err
}
