package document

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Document : interface for a MongoDB document.
type Document interface {
	IsSaved() bool
	Save(ctx context.Context, db *mongo.Database) (string, error)
	Update(ctx context.Context, update interface{}, db *mongo.Database) error
	Delete(ctx context.Context, db *mongo.Database) error
	GetByID(ctx context.Context, id string, db *mongo.Database) error
	SetID(id string) error
	GetID() string
}
