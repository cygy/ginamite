package functions

import (
	"errors"

	"github.com/cygy/ginamite/common/queue"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// RegisterByEmailAddress : saves the registration details of an user.
func RegisterByEmailAddress(c *gin.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	mongoSession := session.Get(c)
	user := model.NewUser()
	return user.RegisterByEmailAddress(username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip, device, mongoSession)
}

// RegisterByThirdPartyToken : saves the registration details of an user.
func RegisterByThirdPartyToken(c *gin.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	mongoSession := session.Get(c)
	user := model.NewUser()
	return user.RegisterByThirdPartyToken(username, tokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip, device, mongoSession)
}

// ValidateUserRegistration : validate the registration of an user.
func ValidateUserRegistration(c *gin.Context, userID, registrationKey string) error {
	// Check the userID.
	if !bson.IsObjectIdHex(userID) {
		return errors.New("wrong object ID")
	}

	mongoSession := session.Get(c)

	// Do not validate or cancel if the account is verified already.
	user := &model.User{}
	if err := user.GetByID(userID, mongoSession); err != nil {
		return err
	}

	if !user.IsSaved() || user.PrivateInfos.IsEmailValid {
		return errors.New("user is already verified")
	}

	if err := user.ValidateRegistrationByEmailAddress(registrationKey, mongoSession); err != nil {
		return err
	}

	return nil
}

// CancelUserRegistration : cancel the registration of an user.
func CancelUserRegistration(c *gin.Context, userID, registrationKey string) error {
	// Check the userID.
	if !bson.IsObjectIdHex(userID) {
		return errors.New("wrong object ID")
	}

	mongoSession := session.Get(c)

	// Do not validate or cancel if the account is verified already.
	user := &model.User{}
	if err := user.GetByID(userID, mongoSession); err != nil {
		return err
	}

	if !user.IsSaved() || user.PrivateInfos.IsEmailValid {
		return errors.New("user is already verified")
	}

	if err := user.CancelRegistrationByEmailAddress(registrationKey, mongoSession); err != nil {
		return err
	}

	queue.DeleteUser(queue.TaskDeleteUser{
		UserID: userID,
	})

	return nil
}

// GetRegistrationDetails : returns the details of the registration of an user.
func GetRegistrationDetails(c *gin.Context, userID string) (*authentication.Details, error) {
	// Check the userID.
	if !bson.IsObjectIdHex(userID) {
		return nil, errors.New("wrong object ID")
	}

	mongoSession := session.Get(c)

	// Do not validate or cancel if the account is enabled already.
	user := &model.User{}
	if err := user.GetByID(userID, mongoSession); err != nil {
		return nil, err
	}

	return &authentication.Details{
		Device: *user.RegistrationInfos.Device.ToDevice(),
		Source: user.RegistrationInfos.Source,
		Method: user.RegistrationInfos.Method,
	}, nil
}
