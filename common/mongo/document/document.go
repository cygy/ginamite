// Package document provides generic operations on the MongoDB documents.
package document

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Document is a MongoDB document, able to save, update, delete and load itself.
type Document interface {
	IsSaved() bool
	Save(ctx context.Context, db *mongo.Database) (string, error)
	Update(ctx context.Context, update any, db *mongo.Database) error
	Delete(ctx context.Context, db *mongo.Database) error
	GetByID(ctx context.Context, id string, db *mongo.Database) error
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
