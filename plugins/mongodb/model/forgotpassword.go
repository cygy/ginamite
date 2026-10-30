package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// ForgotPasswordProcessCollection is the name of the collection of the forgotPassword documents.
const ForgotPasswordProcessCollection = "forgotPassword"

// ForgotPasswordProcess is a pending reset of the password of a user.
type ForgotPasswordProcess struct {
	ID     bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID bson.ObjectID `json:"user_id" bson:"userId"`
	Code   string        `json:"code" bson:"code"`
	Tries  uint          `json:"-" bson:"tries"`
}

// NewForgotPasswordProcess returns a new ForgotPasswordProcess.
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

// IsSaved reports whether the ForgotPasswordProcess document is saved in the collection.
func (fpp *ForgotPasswordProcess) IsSaved() bool {
	return bson.IsObjectIDProvided(fpp.ID)
}

// Save inserts the ForgotPasswordProcess document in the collection, returning an error if needed.
func (fpp *ForgotPasswordProcess) Save(ctx context.Context, db store.Database) (string, error) {
	return document.SaveDocument(ctx, fpp, ForgotPasswordProcessCollection, db)
}

// Update updates the ForgotPasswordProcess document, returning an error if needed.
func (fpp *ForgotPasswordProcess) Update(ctx context.Context, update any, db store.Database) error {
	return document.UpdateDocument(ctx, fpp.ID, ForgotPasswordProcessCollection, update, db)
}

// Delete deletes the ForgotPasswordProcess document from the collection, returning an error if needed.
func (fpp *ForgotPasswordProcess) Delete(ctx context.Context, db store.Database) error {
	return document.DeleteDocument(ctx, fpp.ID, ForgotPasswordProcessCollection, db)
}

// GetByID loads the ForgotPasswordProcess document with the ID.
func (fpp *ForgotPasswordProcess) GetByID(ctx context.Context, id string, db store.Database) error {
	return document.GetDocumentByID(ctx, fpp, ForgotPasswordProcessCollection, id, db)
}

// GetID returns the ForgotPasswordProcess ID.
func (fpp *ForgotPasswordProcess) GetID() string {
	return fpp.ID.Hex()
}

// SetID sets the ID of the ForgotPasswordProcess from its hexadecimal form. The ID is kept if the form is not valid.
func (fpp *ForgotPasswordProcess) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	fpp.ID = objectID
	return nil
}

// GetByUserID loads the process of the user.
func (fpp *ForgotPasswordProcess) GetByUserID(ctx context.Context, userID string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return document.GetDocumentBySelector(ctx, fpp, ForgotPasswordProcessCollection, bson.M{"userId": objectID}, db)
}

// Increment increments the number of tries of the process.
func (fpp *ForgotPasswordProcess) Increment(ctx context.Context, db store.Database) error {
	return fpp.Update(ctx, bson.M{"$inc": bson.M{"tries": 1}}, db)
}

// DeleteForgotPasswordProcessByUserID deletes all the documents linked to a user.
func DeleteForgotPasswordProcessByUserID(ctx context.Context, userID string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = document.DeleteDocumentsBySelector(ctx, ForgotPasswordProcessCollection, bson.M{"userId": objectID}, db)
	return err
}
