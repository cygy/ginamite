package functions

import (
	"errors"

	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/timezone"

	"github.com/cygy/ginamite/common/queue"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/gin-gonic/gin"
)

// registrationTimezone returns the timezone of the user who registers with the locale.
func registrationTimezone(c *gin.Context, locale string) string {
	return model.RegistrationTimezone(locale, timezone.Main, ctxconfig.Get(c).DefaultTimezone)
}

// RegisterByEmailAddress saves the registration details of a user.
func RegisterByEmailAddress(c *gin.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	mongoSession := session.Get(c)
	user := model.NewUser()
	return user.RegisterByEmailAddress(c, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip, device, registrationTimezone(c, locale), mongoSession)
}

// RegisterByThirdPartyToken saves the registration details of a user.
func RegisterByThirdPartyToken(c *gin.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	mongoSession := session.Get(c)
	user := model.NewUser()
	return user.RegisterByThirdPartyToken(c, username, tokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip, device, registrationTimezone(c, locale), mongoSession)
}

// ValidateUserRegistration validates the registration of a user.
func ValidateUserRegistration(c *gin.Context, userID, registrationKey string) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	mongoSession := session.Get(c)

	user, err := unverifiedUser(c, userID)
	if err != nil {
		return err
	}

	return user.ValidateRegistrationByEmailAddress(c, registrationKey, mongoSession)
}

// CancelUserRegistration cancels the registration of a user.
func CancelUserRegistration(c *gin.Context, userID, registrationKey string) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	mongoSession := session.Get(c)

	user, err := unverifiedUser(c, userID)
	if err != nil {
		return err
	}

	if err := user.CancelRegistrationByEmailAddress(c, registrationKey, mongoSession); err != nil {
		return err
	}

	queue.DeleteUser(queue.TaskDeleteUser{
		UserID: userID,
	})

	return nil
}

// GetRegistrationDetails returns the details of the registration of a user.
func GetRegistrationDetails(c *gin.Context, userID string) (*authentication.Details, error) {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return nil, err
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

// unverifiedUser loads the user, and returns an error if its registration is already validated.
func unverifiedUser(c *gin.Context, userID string) (*model.User, error) {
	user := &model.User{}
	if err := user.GetByID(c, userID, session.Get(c)); err != nil {
		return nil, err
	}

	if !user.IsSaved() || user.PrivateInfos.IsEmailValid {
		return nil, errors.New("user is already verified")
	}

	return user, nil
}
