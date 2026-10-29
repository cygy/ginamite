package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// DeletedUser : deleted user properties.
type DeletedUser struct {
	ID        bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    bson.ObjectID `json:"-" bson:"userId,omitempty"`
	CreatedAt time.Time     `json:"created_at" bson:"createdAt"`
}

// NewDeletedUser : returns a new 'DeletedUser' struct.
func NewDeletedUser(userID string) (*DeletedUser, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.WrongObjectID()
	}
	return &DeletedUser{
		UserID:    objectID,
		CreatedAt: time.Now(),
	}, nil
}

// IsSaved : returns true if the 'DeletedUser' document if saved in the collection.
func (deletedUser *DeletedUser) IsSaved() bool {
	return !deletedUser.ID.IsZero()
}

// Save : inserts the 'DeletedUser' document in the collection, returns an error if needed.
func (deletedUser *DeletedUser) Save(ctx context.Context, db *mongo.Database) (string, error) {
	return document.SaveDocument(ctx, deletedUser, DeletedUserCollection, db)
}

// Update : updates the 'DeletedUser' document, returns an error if needed.
func (deletedUser *DeletedUser) Update(ctx context.Context, update any, db *mongo.Database) error {
	return document.UpdateDocument(ctx, deletedUser.ID, DeletedUserCollection, update, db)
}

// Delete : deletes the 'DeletedUser' document from the collection, returns an error if needed.
func (deletedUser *DeletedUser) Delete(ctx context.Context, db *mongo.Database) error {
	return document.DeleteDocument(ctx, deletedUser.ID, DeletedUserCollection, db)
}

// GetByID : initializes the 'DeletedUser' struct from the 'DeletedUser' document with the ID.
func (deletedUser *DeletedUser) GetByID(ctx context.Context, id string, db *mongo.Database) error {
	return document.GetDocumentByID(ctx, deletedUser, DeletedUserCollection, id, db)
}

// GetID : returns the 'DeletedUser' ID.
func (deletedUser *DeletedUser) GetID() string {
	return deletedUser.ID.Hex()
}

// SetID : initializes the 'DeletedUser' ID.
func (deletedUser *DeletedUser) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.WrongObjectID()
	}
	deletedUser.ID = objectID
	return nil
}
