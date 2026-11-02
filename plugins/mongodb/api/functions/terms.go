package functions

import (
	"context"

	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// Terms stores the version of the terms accepted by the users in a database. It is a terms.Storage. Create it with
// NewTerms.
type Terms struct {
	database func() store.Database
}

// NewTerms returns the storage of the terms working on the database returned by the function, which is called each
// time the storage needs the database.
func NewTerms(database func() store.Database) *Terms {
	return &Terms{database: database}
}

// UpdateVersionOfTermsAcceptedByUser saves the version of the terms accepted by the user.
func (s *Terms) UpdateVersionOfTermsAcceptedByUser(ctx context.Context, userID, version string) error {
	user, err := model.UserWithID(userID)
	if err != nil {
		return err
	}

	return user.AcceptVersionOfTerms(ctx, version, s.database())
}

// GetLatestVersionOfTermsAcceptedByUser returns the latest version of the terms accepted by the user, or an empty
// string if the user does not exist.
func (s *Terms) GetLatestVersionOfTermsAcceptedByUser(ctx context.Context, userID string) (string, error) {
	user := &model.User{}
	if err := user.GetByID(ctx, userID, s.database()); err != nil {
		if model.IsNotFound(err) {
			return "", nil
		}

		return "", err
	}

	return user.VersionOfTermsAccepted, nil
}
