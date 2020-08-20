package functions

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/common/authentication"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/api/context"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// GetUserDetailsByIdentifier : returns the details of an user by its identifier (email address or username).
func GetUserDetailsByIdentifier(c *gin.Context, identifier string) (authentication.User, error) {
	mongoSession := session.Get(c)

	// Get the user by the identifier of the email address.
	var err error
	user := &model.User{}
	if strings.Contains(identifier, "@") {
		err = user.GetByEmailAddress(identifier, mongoSession)
	} else {
		err = user.GetByUsername(identifier, mongoSession)
	}

	if err != nil {
		return authentication.User{}, err
	}

	return user.ToUser(), nil
}

// GetUserDetailsByID : returns the details of an user by its ID.
func GetUserDetailsByID(c *gin.Context, userID string) (authentication.User, error) {
	// Check the userID.
	if !bson.IsObjectIdHex(userID) {
		return authentication.User{}, errors.New("wrong object ID")
	}

	cachedUser := context.GetFullUser(c)
	if cachedUser != nil && cachedUser.ID.Hex() == userID {
		return cachedUser.ToUser(), nil
	}

	mongoSession := session.Get(c)

	user := &model.User{}
	if err := user.GetByID(userID, mongoSession); err != nil {
		return authentication.User{}, err
	}

	return user.ToUser(), nil
}

// GetUserDetailsByFacebookUserID : returns the details of an user by its unique facebook ID.
func GetUserDetailsByFacebookUserID(c *gin.Context, userID string) (authentication.User, error) {
	mongoSession := session.Get(c)
	user := &model.User{}
	err := user.GetByFacebookUserID(userID, mongoSession)

	return user.ToUser(), err
}

// GetUserDetailsByGoogleUserID : returns the details of an user by its unique google ID.
func GetUserDetailsByGoogleUserID(c *gin.Context, userID string) (authentication.User, error) {
	mongoSession := session.Get(c)
	user := &model.User{}
	err := user.GetByGoogleUserID(userID, mongoSession)

	return user.ToUser(), err
}

// GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser : returns the details of an user and its encrypted password by its identifier (email address or username).
func GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(c *gin.Context, identifier string) (authentication.User, string, error) {
	mongoSession := session.Get(c)
	user := &model.User{}
	if err := user.GetByIdentifier(identifier, mongoSession); err != nil {
		return authentication.User{}, "", err
	}

	if !user.IsSaved() {
		return authentication.User{}, "", commonErrors.NotFound()
	}

	return user.ToUser(), user.IdentificationMethods.Password, nil
}

// DeleteUserByID : deletes an user by its ID.
func DeleteUserByID(c *gin.Context, userID string) error {
	mongoSession := session.Get(c)

	deletedUser := model.NewDeletedUser(userID)
	if _, err := deletedUser.Save(mongoSession); err != nil {
		return err
	}

	user := model.UserWithID(userID)

	return user.Delete(mongoSession)
}

// DisableUserByID : disables an user by its ID.
func DisableUserByID(c *gin.Context, userID string) (authentication.User, error) {
	mongoSession := session.Get(c)

	user := &model.User{}
	if err := user.GetByID(userID, mongoSession); err != nil {
		return authentication.User{}, err
	}

	disabledUser := model.NewDisabledUser(user)
	if _, err := disabledUser.Save(mongoSession); err != nil {
		return authentication.User{}, err
	}

	return user.ToUser(), user.Delete(mongoSession)
}

// EnableUserByID : enables an user by its ID.
func EnableUserByID(c *gin.Context, userID string) (authentication.User, error) {
	mongoSession := session.Get(c)

	disabledUser := &model.DisabledUser{}
	if err := disabledUser.GetByID(userID, mongoSession); err != nil {
		return authentication.User{}, err
	}

	user := model.NewUserFromDisabledUser(disabledUser)
	if _, err := user.Save(mongoSession); err != nil {
		return authentication.User{}, err
	}

	return user.ToUser(), disabledUser.Delete(mongoSession)
}
