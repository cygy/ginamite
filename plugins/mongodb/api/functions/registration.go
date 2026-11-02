package functions

import (
	"context"
	"errors"

	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
)

// RegisterByEmailAddress saves the registration details of a user.
func (a *Accounts) RegisterByEmailAddress(ctx context.Context, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	mongoSession := a.database()
	user := model.NewUser()
	return user.RegisterByEmailAddress(ctx, username, encryptedPassword, emailAddress, locale, termsVersion, registrationCode, privateKey, source, ip, device, a.registrationDefaults(locale), mongoSession)
}

// RegisterByThirdPartyToken saves the registration details of a user.
func (a *Accounts) RegisterByThirdPartyToken(ctx context.Context, username string, tokenInfos validator.TokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip string, device authentication.Device) error {
	mongoSession := a.database()
	user := model.NewUser()
	return user.RegisterByThirdPartyToken(ctx, username, tokenInfos, tokenSource, locale, termsVersion, registrationCode, privateKey, source, ip, device, a.registrationDefaults(locale), mongoSession)
}

// ValidateUserRegistration validates the registration of a user.
func (a *Accounts) ValidateUserRegistration(ctx context.Context, userID, registrationKey string) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	mongoSession := a.database()

	user, err := a.unverifiedUser(ctx, userID)
	if err != nil {
		return err
	}

	return user.ValidateRegistrationByEmailAddress(ctx, registrationKey, a.settings().UserMaxValidateRegistrationAttemps, mongoSession)
}

// CancelUserRegistration cancels the registration of a user.
func (a *Accounts) CancelUserRegistration(ctx context.Context, userID, registrationKey string) error {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return err
	}

	mongoSession := a.database()

	user, err := a.unverifiedUser(ctx, userID)
	if err != nil {
		return err
	}

	if err := user.CancelRegistrationByEmailAddress(ctx, registrationKey, mongoSession); err != nil {
		return err
	}

	a.deleteUser(userID)

	return nil
}

// GetRegistrationDetails returns the details of the registration of a user.
func (a *Accounts) GetRegistrationDetails(ctx context.Context, userID string) (*authentication.Details, error) {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return nil, err
	}

	mongoSession := a.database()

	// Do not validate or cancel if the account is enabled already.
	user := &model.User{}
	if err := user.GetByID(ctx, userID, mongoSession); err != nil {
		return nil, err
	}

	return &authentication.Details{
		Device: *user.RegistrationInfos.Device.ToDevice(),
		Source: user.RegistrationInfos.Source,
		Method: user.RegistrationInfos.Method,
	}, nil
}

// unverifiedUser loads the user, and returns an error if its registration is already validated.
func (a *Accounts) unverifiedUser(ctx context.Context, userID string) (*model.User, error) {
	user := &model.User{}
	if err := user.GetByID(ctx, userID, a.database()); err != nil {
		return nil, err
	}

	if !user.IsSaved() || user.PrivateInfos.IsEmailValid {
		return nil, errors.New("user is already verified")
	}

	return user, nil
}
