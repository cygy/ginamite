package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AcceptVersionOfTerms : updates the version of the terms accepted.
func (user *User) AcceptVersionOfTerms(ctx context.Context, version string, db store.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"versionOfTermsAccepted": version}}, db)
}
