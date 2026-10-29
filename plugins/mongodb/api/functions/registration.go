package functions

import (
	"errors"

	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/queue"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// RegisterByEmailAddress : saves the registration details of an user.
func RegisterByEmailAddress(c *gin.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	mongoSession := session.Get(c)
	user := model.NewUser()
	return user.RegisterByEmailAddress(c, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip, device, mongoSession)
}

// RegisterByThirdPartyToken : saves the registration details of an user.
func RegisterByThirdPartyToken(c *gin.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	mongoSession := session.Get(c)
	user := model.NewUser()
	return user.RegisterByThirdPartyToken(c, username, tokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip, device, mongoSession)
}

// ValidateUserRegistration : validate the registration of an user.
func ValidateUserRegistration(c *gin.Context, userID, registrationKey string) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return commonErrors.WrongObjectID()
	}

	mongoSession := session.Get(c)

	// Do not validate or cancel if the account is verified already.
	user := &model.User{}
	if err := user.GetByID(c, userID, mongoSession); err != nil {
		return err
	}

	if !user.IsSaved() || user.PrivateInfos.IsEmailValid {
		return errors.New("user is already verified")
	}

	if err := user.ValidateRegistrationByEmailAddress(c, registrationKey, mongoSession); err != nil {
		return err
	}

	return nil
}

// CancelUserRegistration : cancel the registration of an user.
func CancelUserRegistration(c *gin.Context, userID, registrationKey string) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return commonErrors.WrongObjectID()
	}

	mongoSession := session.Get(c)

	// Do not validate or cancel if the account is verified already.
	user := &model.User{}
	if err := user.GetByID(c, userID, mongoSession); err != nil {
		return err
	}

	if !user.IsSaved() || user.PrivateInfos.IsEmailValid {
		return errors.New("user is already verified")
	}

	if err := user.CancelRegistrationByEmailAddress(c, registrationKey, mongoSession); err != nil {
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
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return nil, commonErrors.WrongObjectID()
	}

	mongoSession := session.Get(c)

	// Do not validate or cancel if the account is enabled already.
	user := &model.User{}
	if err := user.GetByID(c, userID, mongoSession); err != nil {
		return nil, err
	}

	return &authentication.Details{
		Device: *user.RegistrationInfos.Device.ToDevice(),
		Source: user.RegistrationInfos.Source,
		Method: user.RegistrationInfos.Method,
	}, nil
}
