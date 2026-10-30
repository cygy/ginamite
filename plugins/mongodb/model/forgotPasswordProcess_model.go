package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// ForgotPasswordProcess : details about the password to update.
type ForgotPasswordProcess struct {
	ID     bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID bson.ObjectID `json:"user_id" bson:"userId"`
	Code   string        `json:"code" bson:"code"`
	Tries  uint          `json:"-" bson:"tries"`
}

// NewForgotPasswordProcess : returns a new 'ForgotPasswordProcess' struct.
func NewForgotPasswordProcess(userID string) (*ForgotPasswordProcess, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return &ForgotPasswordProcess{
		UserID: objectID,
		Tries:  0,
	}, nil
}

// IsSaved : returns true if the 'ForgotPasswordProcess' document if saved in the collection.
func (fpp *ForgotPasswordProcess) IsSaved() bool {
	return bson.IsObjectIDProvided(fpp.ID)
}

// Save : inserts the 'ForgotPasswordProcess' document in the collection, returns an error if needed.
func (fpp *ForgotPasswordProcess) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, fpp, ForgotPasswordProcessCollection, db)
}

// Update : updates the 'ForgotPasswordProcess' document, returns an error if needed.
func (fpp *ForgotPasswordProcess) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, fpp.ID, ForgotPasswordProcessCollection, update, db)
}

// Delete : deletes the 'ForgotPasswordProcess' document from the collection, returns an error if needed.
func (fpp *ForgotPasswordProcess) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, fpp.ID, ForgotPasswordProcessCollection, db)
}

// GetByID : initializes the 'ForgotPasswordProcess' struct from the 'user' document with the ID.
func (fpp *ForgotPasswordProcess) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, fpp, ForgotPasswordProcessCollection, id, db)
}

// GetID : returns the 'ForgotPasswordProcess' ID.
func (fpp *ForgotPasswordProcess) GetID() string {
	return fpp.ID.Hex()
}

// SetID : initializes the 'ForgotPasswordProcess' ID.
func (fpp *ForgotPasswordProcess) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	fpp.ID = objectID
	return nil
}
