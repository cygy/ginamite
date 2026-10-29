package model

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AcceptVersionOfTerms : updates the version of the terms accepted.
func (user *User) AcceptVersionOfTerms(ctx context.Context, version string, db *mongo.Database) error {
	return user.Update(ctx, bson.M{"$set": bson.M{"versionOfTermsAccepted": version}}, db)
}
