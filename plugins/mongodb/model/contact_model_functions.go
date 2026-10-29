package model

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UpdateDone : updates the 'done' properties of a contact.
func (contact *Contact) UpdateDone(ctx context.Context, done bool, db *mongo.Database) error {
	return contact.Update(ctx, bson.M{"$set": bson.M{"done": done}}, db)
}
