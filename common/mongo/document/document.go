// Package document provides generic operations on the MongoDB documents.
package document

import (
	"context"
	"strings"

	"github.com/cygy/ginamite/common/mongo/store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Document is a MongoDB document, able to save, update, delete and load itself.
type Document interface {
	IsSaved() bool
	Save(ctx context.Context, db store.Database) (string, error)
	Update(ctx context.Context, update any, db store.Database) error
	Delete(ctx context.Context, db store.Database) error
	GetByID(ctx context.Context, id string, db store.Database) error
	SetID(id string) error
	GetID() string
}

// sortKeys converts mgo-style sort fields ("name", "-created_at") to an ordered bson.D.
func sortKeys(fields ...string) bson.D {
	keys := make(bson.D, 0, len(fields))
	for _, field := range fields {
		value := 1
		if strings.HasPrefix(field, "-") {
			field, value = field[1:], -1
		}
		keys = append(keys, bson.E{Key: field, Value: value})
	}
	return keys
}
