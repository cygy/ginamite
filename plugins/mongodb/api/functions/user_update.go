package functions

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/model"
)

// UpdateUserPassword updates the password of a user.
func (a *Accounts) UpdateUserPassword(ctx context.Context, userID, password string, isPasswordEncrypted bool) error {
	mongoSession := a.database()
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdatePassword(ctx, password, isPasswordEncrypted, mongoSession)
}

// UpdateUserEmailAddress updates the email address of a user.
func (a *Accounts) UpdateUserEmailAddress(ctx context.Context, userID, emailAddress string) error {
	mongoSession := a.database()
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateEmailAddress(ctx, emailAddress, mongoSession)
}

// UserEmailAddressIsVerified marks the email address of the user as verified.
func (a *Accounts) UserEmailAddressIsVerified(ctx context.Context, userID string) error {
	mongoSession := a.database()
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.SetEmailAddressValid(ctx, mongoSession)
}

// UpdateUserSocialNetworks updates the social network links of a user.
func (a *Accounts) UpdateUserSocialNetworks(ctx context.Context, userID, facebook, twitter, instagram string) error {
	mongoSession := a.database()
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.UpdateSocialNetworks(ctx, facebook, twitter, instagram, mongoSession)
}
