package functions

import (
	"context"
	"strings"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// Accounts are the functions of the accounts of the users working on a MongoDB database.
type Accounts struct {
	database func() store.Database
	settings func() model.Settings

	// registrationDefaults returns the values given to a user who registers with the locale.
	registrationDefaults func(locale string) model.RegistrationDefaults

	// deleteUser asks the workers to delete the data of a user.
	deleteUser func(userID string)
}

// NewAccounts returns the functions of the accounts working on the database returned by the function, which is called
// each time a function needs the database. The function registrationDefaults returns the values given to a user who
// registers with the locale, and deleteUser asks the workers to delete the data of a user.
func NewAccounts(database func() store.Database, settings func() model.Settings, registrationDefaults func(locale string) model.RegistrationDefaults, deleteUser func(userID string)) *Accounts {
	return &Accounts{database: database, settings: settings, registrationDefaults: registrationDefaults, deleteUser: deleteUser}
}

// GetUserDetailsByIdentifier returns the details of a user by its identifier (email address or username).
func (a *Accounts) GetUserDetailsByIdentifier(ctx context.Context, identifier string) (authentication.User, error) {
	mongoSession := a.database()

	// Get the user by the identifier of the email address.
	var err error
	user := &model.User{}
	if strings.Contains(identifier, "@") {
		err = user.GetByEmailAddress(ctx, identifier, mongoSession)
	} else {
		err = user.GetByUsername(ctx, identifier, mongoSession)
	}

	if err != nil {
		return authentication.User{}, err
	}

	return user.ToUser(), nil
}

// GetUserDetailsByID returns the details of a user by its ID.
func (a *Accounts) GetUserDetailsByID(ctx context.Context, userID string) (authentication.User, error) {
	// Check the userID.
	if _, err := bson.ObjectIDFromHex(userID); err != nil {
		return authentication.User{}, err
	}

	mongoSession := a.database()

	user := &model.User{}
	if err := user.GetByID(ctx, userID, mongoSession); err != nil {
		return authentication.User{}, err
	}

	return user.ToUser(), nil
}

// GetUserDetailsByFacebookUserID returns the details of a user by its unique facebook ID.
func (a *Accounts) GetUserDetailsByFacebookUserID(ctx context.Context, userID string) (authentication.User, error) {
	mongoSession := a.database()
	user := &model.User{}
	err := user.GetByFacebookUserID(ctx, userID, mongoSession)

	return user.ToUser(), err
}

// GetUserDetailsByGoogleUserID returns the details of a user by its unique google ID.
func (a *Accounts) GetUserDetailsByGoogleUserID(ctx context.Context, userID string) (authentication.User, error) {
	mongoSession := a.database()
	user := &model.User{}
	err := user.GetByGoogleUserID(ctx, userID, mongoSession)

	return user.ToUser(), err
}

// GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser returns the details and the encrypted password of an
// enabled user, by its identifier: the username or the email address.
func (a *Accounts) GetUserDetailsAndEncryptedPasswordByIdentifierFromEnabledUser(ctx context.Context, identifier string) (authentication.User, string, error) {
	mongoSession := a.database()
	user := &model.User{}
	if err := user.GetByIdentifier(ctx, identifier, mongoSession); err != nil {
		return authentication.User{}, "", err
	}

	if !user.IsSaved() {
		return authentication.User{}, "", errors.NotFound()
	}

	return user.ToUser(), user.IdentificationMethods.Password, nil
}

// UpdateEncryptedPasswordByUserID replaces the encrypted password of a user.
func (a *Accounts) UpdateEncryptedPasswordByUserID(ctx context.Context, userID, encryptedPassword string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdatePassword(ctx, encryptedPassword, true, a.database())
}

// DeleteUserByID deletes a user by its ID.
func (a *Accounts) DeleteUserByID(ctx context.Context, userID string) error {
	mongoSession := a.database()

	deletedUser, err := model.NewDeletedUser(userID)
	if err != nil {
		return err
	}
	if _, err = deletedUser.Save(ctx, mongoSession); err != nil {
		return err
	}

	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.Delete(ctx, mongoSession)
}

// DisableUserByID moves the user to the disabled users and returns its details.
func (a *Accounts) DisableUserByID(ctx context.Context, userID string) (authentication.User, error) {
	mongoSession := a.database()

	user := &model.User{}
	if err := user.GetByID(ctx, userID, mongoSession); err != nil {
		return authentication.User{}, err
	}

	disabledUser := model.NewDisabledUser(user)
	if _, err := disabledUser.Save(ctx, mongoSession); err != nil {
		return authentication.User{}, err
	}

	return user.ToUser(), user.Delete(ctx, mongoSession)
}

// EnableUserByID restores the disabled user as an enabled user and returns its details.
func (a *Accounts) EnableUserByID(ctx context.Context, userID string) (authentication.User, error) {
	mongoSession := a.database()

	disabledUser := &model.DisabledUser{}
	if err := disabledUser.GetByID(ctx, userID, mongoSession); err != nil {
		return authentication.User{}, err
	}

	user := model.NewUserFromDisabledUser(disabledUser)
	if _, err := user.Save(ctx, mongoSession); err != nil {
		return authentication.User{}, err
	}

	return user.ToUser(), disabledUser.Delete(ctx, mongoSession)
}
