package model

import (
	"context"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
)

// DeletedUserCollection is the name of the collection of the deletedUsers documents.
const DeletedUserCollection = "deletedUsers"

// DeletedUser records that a user deleted their account.
type DeletedUser struct {
	ID        bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    bson.ObjectID `json:"-" bson:"userId,omitempty"`
	CreatedAt time.Time     `json:"created_at" bson:"createdAt"`
}

// NewDeletedUser returns a new DeletedUser.
func NewDeletedUser(userID string) (*DeletedUser, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	return &DeletedUser{
		UserID:    objectID,
		CreatedAt: time.Now(),
	}, nil
}

// IsSaved reports whether the DeletedUser document is saved in the collection.
func (deletedUser *DeletedUser) IsSaved() bool {
	return bson.IsObjectIDProvided(deletedUser.ID)
}

// Save inserts the DeletedUser document in the collection, returning an error if needed.
func (deletedUser *DeletedUser) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, deletedUser, DeletedUserCollection, db)
}

// Update updates the DeletedUser document, returning an error if needed.
func (deletedUser *DeletedUser) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, deletedUser.ID, DeletedUserCollection, update, db)
}

// Delete deletes the DeletedUser document from the collection, returning an error if needed.
func (deletedUser *DeletedUser) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, deletedUser.ID, DeletedUserCollection, db)
}

// GetByID loads the DeletedUser document with the ID.
func (deletedUser *DeletedUser) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, deletedUser, DeletedUserCollection, id, db)
}

// GetID returns the DeletedUser ID.
func (deletedUser *DeletedUser) GetID() string {
	return deletedUser.ID.Hex()
}

// SetID sets the ID of the DeletedUser from its hexadecimal form. The ID is kept if the form is not valid.
func (deletedUser *DeletedUser) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	deletedUser.ID = objectID
	return nil
}

// DeleteDeletedUserByUserID deletes the document with the ID.
func DeleteDeletedUserByUserID(ctx context.Context, userID string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = document.DeleteDocumentsBySelector(ctx, DeletedUserCollection, bson.M{"userId": objectID}, db)
	return err
}
