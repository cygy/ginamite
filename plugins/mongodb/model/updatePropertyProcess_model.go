package model

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/random"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// UpdatePropertyProcess : details about a property to update.
type UpdatePropertyProcess struct {
	ID     bson.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID bson.ObjectID `json:"user_id" bson:"userId"`
	Type   string        `json:"type" bson:"type"`
	Value  string        `json:"value" bson:"value"`
	Code   string        `json:"code" bson:"code"`
	Tries  uint          `json:"-" bson:"tries"`
	TTLKey string        `json:"-" bson:"userIdtype"` // This field is indexed with a TTL. A TTL index must use only one field, so the pair 'userId / type' can not be used.
}

// NewUpdatePropertyProcess : returns a new 'UpdatePropertyProcess' struct.
func NewUpdatePropertyProcess(userID, t, value string) (*UpdatePropertyProcess, error) {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return &UpdatePropertyProcess{
		UserID: objectID,
		Type:   t,
		Value:  value,
		Code:   random.String(64),
		Tries:  0,
		TTLKey: fmt.Sprintf("%s%s", userID, t),
	}, nil
}

// NewUpdatePasswordProcess : returns a new 'UpdatePropertyProcess' struct.
func NewUpdatePasswordProcess(userID, value string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessUpdatePassword, value)
}

// NewUpdateEmailAddressProcess : returns a new 'UpdatePropertyProcess' struct.
func NewUpdateEmailAddressProcess(userID, value string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessUpdateEmailAddress, value)
}

// NewUpdateValidEmailAddressProcess : returns a new 'UpdatePropertyProcess' struct.
func NewUpdateValidEmailAddressProcess(userID string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessVerifyEmailAddress, "true")
}

// NewDeleteAccountProcess : returns a new 'UpdatePropertyProcess' struct.
func NewDeleteAccountProcess(userID string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessDeleteAccount, "true")
}

// NewDisableAccountProcess : returns a new 'UpdatePropertyProcess' struct.
func NewDisableAccountProcess(userID string) (*UpdatePropertyProcess, error) {
	return NewUpdatePropertyProcess(userID, authentication.ProcessDisableAccount, "true")
}

// IsSaved : returns true if the 'UpdatePropertyProcess' document if saved in the collection.
func (process *UpdatePropertyProcess) IsSaved() bool {
	return bson.IsObjectIDProvided(process.ID)
}

// Save : inserts the 'UpdatePropertyProcess' document in the collection, returns an error if needed.
func (process *UpdatePropertyProcess) Save(ctx context.Context, db *mongo.Database) (string, error) {
	return document.SaveDocument(ctx, process, UpdatePropertyProcessCollection, db)
}

// Update : updates the 'UpdatePropertyProcess' document, returns an error if needed.
func (process *UpdatePropertyProcess) Update(ctx context.Context, update any, db *mongo.Database) error {
	return document.UpdateDocument(ctx, process.ID, UpdatePropertyProcessCollection, update, db)
}

// Delete : deletes the 'UpdatePropertyProcess' document from the collection, returns an error if needed.
func (process *UpdatePropertyProcess) Delete(ctx context.Context, db *mongo.Database) error {
	return document.DeleteDocument(ctx, process.ID, UpdatePropertyProcessCollection, db)
}

// GetByID : initializes the 'UpdatePropertyProcess' struct from the 'user' document with the ID.
func (process *UpdatePropertyProcess) GetByID(ctx context.Context, id string, db *mongo.Database) error {
	return document.GetDocumentByID(ctx, process, UpdatePropertyProcessCollection, id, db)
}

// GetID : returns the 'UpdatePropertyProcess' ID.
func (process *UpdatePropertyProcess) GetID() string {
	return process.ID.Hex()
}

// SetID : initializes the 'UpdatePropertyProcess' ID.
func (process *UpdatePropertyProcess) SetID(id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	process.ID = objectID
	return nil
}
