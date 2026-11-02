package functions

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// DoesAccountWithUsernameExist reports whether an account with this username exists.
func (a *Accounts) DoesAccountWithUsernameExist(ctx context.Context, username string) (bool, error) {
	return a.existsInEither(ctx, username, model.IsUserWithUsernameExisting, model.IsDisabledUserWithUsernameExisting)
}

// DoesAccountWithEmailAddressExist reports whether an account with this email address exists.
func (a *Accounts) DoesAccountWithEmailAddressExist(ctx context.Context, emailAddress string) (bool, error) {
	return a.existsInEither(ctx, emailAddress, model.IsUserWithEmailAddressExisting, model.IsDisabledUserWithEmailAddressExisting)
}

// DoesUserWithFacebookUserIDExist reports whether an account with this facebook id exists.
func (a *Accounts) DoesUserWithFacebookUserIDExist(ctx context.Context, userID string) (bool, error) {
	return a.existsInEither(ctx, userID, model.IsUserWithFacebookUserIDExisting, model.IsDisabledUserWithFacebookUserIDExisting)
}

// DoesUserWithGoogleUserIDExist reports whether an account with this google id exists.
func (a *Accounts) DoesUserWithGoogleUserIDExist(ctx context.Context, userID string) (bool, error) {
	return a.existsInEither(ctx, userID, model.IsUserWithGoogleUserIDExisting, model.IsDisabledUserWithGoogleUserIDExisting)
}

// existsInEither reports whether the value is found by the check of the enabled users or, if not, by the check of
// the disabled users.
func (a *Accounts) existsInEither(ctx context.Context, value string, enabled, disabled func(context.Context, string, store.Database) (bool, error)) (bool, error) {
	mongoSession := a.database()

	ok, err := enabled(ctx, value, mongoSession)
	if err != nil {
		return false, err
	}

	if !ok {
		ok, err = disabled(ctx, value, mongoSession)
	}

	return ok, err
}
