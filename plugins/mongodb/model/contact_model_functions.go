package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UpdateDone : updates the 'done' properties of a contact.
func (contact *Contact) UpdateDone(ctx context.Context, done bool, db store.Database) error {
	return contact.Update(ctx, bson.M{"$set": bson.M{"done": done}}, db)
}
